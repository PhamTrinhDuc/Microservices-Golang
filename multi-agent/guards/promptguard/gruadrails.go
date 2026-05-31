package promptguard

import (
	"context"
	"sync"
)

type Guardrails struct {
	inputGuards  []Guard
	outputGuards []Guard
	parallel     bool // true: execute guard input/output parallel
	failFast     bool // true: stop execute when found first unsafe guard
}

type Option func(*Guardrails)

func WithParallel() Option { return func(g *Guardrails) { g.parallel = true } }
func WithFailFast() Option { return func(g *Guardrails) { g.failFast = true } }

func New(opts ...Option) *Guardrails {
	g := &Guardrails{failFast: true} // default: fail fast
	for _, o := range opts {
		o(g)
	}
	return g
}

// AddInputGuard: add guards to list inputGuards to process input sequence/parallel
func (g *Guardrails) AddInputGuard(guards ...Guard) *Guardrails {
	g.inputGuards = append(g.inputGuards, guards...)
	return g
}

// AddOutputGuard: add guards to list outputGuards to process output sequence/parallel
func (g *Guardrails) AddOutputGuard(guards ...Guard) *Guardrails {
	g.outputGuards = append(g.outputGuards, guards...)
	return g
}

// runSequential present method to run guards sequentially, stop when found first unsafe guard if failFast true
func (g *Guardrails) runSequential(ctx context.Context, guards []Guard, input, output string, isInput bool) (*GuardResult, error) {
	result := &GuardResult{}

	for _, guard := range guards {
		if isInput && !guard.SupportInput() {
			continue
		}
		if !isInput && !guard.SupportOutput() {
			continue
		}

		var (
			verdict GuardVerdict
			err     error
		)

		if isInput {
			verdict, err = guard.GuardInput(ctx, input)
		} else {
			verdict, err = guard.GuardOutput(ctx, input, output)
		}

		if err != nil {
			return nil, err
		}

		result.Verdicts = append(result.Verdicts, verdict)

		if verdict.Verdict == VerdictUnsafe || verdict.Verdict == VerdictUncertain {
			result.Breached = true
			if g.failFast {
				return result, nil
			}
		}
	}

	return result, nil
}

func (g *Guardrails) runParallel(ctx context.Context, guards []Guard, input, output string, isInput bool) (*GuardResult, error) {
	type item struct {
		verdict GuardVerdict
		err     error
	}

	ch := make(chan item, len(guards))
	var wg sync.WaitGroup

	for _, guard := range guards {
		if isInput && !guard.SupportInput() {
			continue
		}
		if !isInput && !guard.SupportOutput() {
			continue
		}

		wg.Add(1)
		go func(g Guard) {
			defer wg.Done()
			var (
				verdict GuardVerdict
				err     error
			)
			if isInput {
				verdict, err = g.GuardInput(ctx, input)
			} else {
				verdict, err = g.GuardOutput(ctx, input, output)
			}

			ch <- item{
				verdict: verdict,
				err:     err,
			}
		}(guard)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	result := &GuardResult{}
	for it := range ch {
		if it.err != nil {
			return nil, it.err
		}
		result.Verdicts = append(result.Verdicts, it.verdict)
		if it.verdict.Verdict == VerdictUnsafe || it.verdict.Verdict == VerdictUncertain {
			result.Breached = true
			if g.failFast {
				return result, nil
			}
		}
	}

	return result, nil
}

// GuardInput present method to guard input from user/developer before send to LLM
func (g *Guardrails) GuardInput(ctx context.Context, input string) (*GuardResult, error) {
	if !g.parallel {
		return g.runSequential(ctx, g.inputGuards, input, "", true)
	}
	return g.runParallel(ctx, g.inputGuards, input, "", true)
}

// GuardOutput present method to guard output from LLM before send to user/developer
func (g *Guardrails) GuardOutput(ctx context.Context, input, output string) (*GuardResult, error) {
	if !g.parallel {
		return g.runSequential(ctx, g.outputGuards, input, output, false)
	}
	return g.runParallel(ctx, g.outputGuards, input, output, false)
}

package promptguard

import "context"

type Verdict string

const (
	VerdictSafe       Verdict = "safe"
	VerdictUnsafe     Verdict = "unsafe"
	VerdictBorderline Verdict = "borderline"
	VerdictUncertain  Verdict = "uncertain"
)

// GuardVerdict represent single guard result
type GuardVerdict struct {
	GuardName string
	Verdict   Verdict
	Reason    string
	Layer     string // rule, slm, cache
	Metadata  map[string]any
}

// GuardResult - present result from all Guards
type GuardResult struct {
	Breached bool
	Verdicts []GuardVerdict
	CacheHit bool
}

// Guard - interface for any guard implement
type Guard interface {
	Name() string
	// true if guard handle input/output, false don't handle
	SupportInput() bool
	SupportOutput() bool
	// guard handle input / output
	GuardInput(ctx context.Context, input string) (GuardVerdict, error)
	GuardOutput(ctx context.Context, input, output string) (GuardVerdict, error)
}

type ThresholdConfig struct {
	// If score >= UnsafeThreshold -> Unsafe
	// If score >= BorderlineThreshold -> Borderline
	UnsafeThreshold     float64
	BorderlineThreshold float64
}

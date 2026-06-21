package rule

import (
	"context"
	"fmt"
	"single-agent/guards/promptguard"
	"regexp"
	"strings"
)

// Rule definition for rule based guard
type Rule struct {
	Name          string   `yaml:"name"`
	Keywords      []string `yaml:"keywords"`
	Pattern       string   `yaml:"pattern"`   // 1 regex string
	Patterns      []string `yaml:"patterns"`  // multiple regex strings
	MatchAll      bool     `yaml:"match_all"` // false = any(default)
	MatchType     string   `yaml:"match_type"` // "any"|"all"
	CaseSensitive bool     `yaml:"case_sensitive"`
	Verdict       string   `yaml:"verdict"`  // safe|unsafe|borderline|uncertain
	ApplyTo       string   `yaml:"apply_to"` // "input"|"output"|"both"
	ApplyToInput  bool     `yaml:"apply_to_input"`
	ApplyToOutput bool     `yaml:"apply_to_output"`
	Tags          []string `yaml:"tags"`
}

// engine hold compiled rules, ready to use
type engine struct {
	rules []compiledRule
}

// compiledRule store compiled regex and flag
type compiledRule struct {
	Rule
	res []*regexp.Regexp // compile from Rule.Pattern and Rule.Patterns
}

func NewEngine(rules []Rule) (*engine, error) {
	e := &engine{}
	for _, r := range rules {
		cr := compiledRule{Rule: r}
		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				return nil, fmt.Errorf("rule %q bad pattern: %w", r.Name, err)
			}
			cr.res = append(cr.res, re)
		}
		for _, pat := range r.Patterns {
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, fmt.Errorf("rule %q bad pattern: %w", r.Name, err)
			}
			cr.res = append(cr.res, re)
		}
		e.rules = append(e.rules, cr)
	}
	return e, nil
}

// matches: check if input text match the rule
func (cr *compiledRule) matches(text string) bool {
	src := text
	var keywords []string
	if !cr.CaseSensitive {
		src = strings.ToLower(text)
		keywords = make([]string, len(cr.Keywords))
		for i, kw := range cr.Keywords {
			keywords[i] = strings.ToLower(kw)
		}
	} else {
		keywords = cr.Keywords
	}

	isMatchAll := cr.MatchAll || cr.MatchType == "all"

	if isMatchAll {
		// AND: match all keywords + pattern
		for _, kw := range keywords {
			if !strings.Contains(src, kw) {
				return false
			}
		}
		for _, re := range cr.res {
			if !re.MatchString(text) {
				return false
			}
		}
		return true
	}
	// OR (default): match any is enough
	for _, kw := range keywords {
		if strings.Contains(src, kw) {
			return true
		}
	}
	for _, re := range cr.res {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func (r *compiledRule) appliesTo(forInput bool) bool {
	if r.ApplyToInput || r.ApplyToOutput {
		if forInput {
			return r.ApplyToInput
		}
		return r.ApplyToOutput
	}

	switch r.ApplyTo {
	case "input":
		return forInput
	case "output":
		return !forInput
	default: // both or unset
		return true
	}
}

func (e *engine) evaluate(text string, forInput bool) *compiledRule {
	for i := range e.rules {
		r := &e.rules[i]
		if !r.appliesTo(forInput) {
			continue
		}
		if r.matches(text) {
			return r
		}
	}
	return nil
}

// implement guard interface for rule based
type Guard struct {
	name   string
	engine *engine
}

func New(name string, rules []Rule) (*Guard, error) {
	e, err := NewEngine(rules)
	if err != nil {
		return nil, err
	}
	return &Guard{
		name:   name,
		engine: e,
	}, nil
}

// implement methods from Guard interface
func (g *Guard) Name() string { return g.name }

func (g *Guard) SupportInput() bool { return true }

func (g *Guard) SupportOutput() bool { return true }

func (g *Guard) GuardInput(ctx context.Context, input string) (promptguard.GuardVerdict, error) {
	return g.check(input, true)
}

func (g *Guard) GuardOutput(ctx context.Context, input, output string) (promptguard.GuardVerdict, error) {
	return g.check(output, false)
}

func (g *Guard) check(text string, forInput bool) (promptguard.GuardVerdict, error) {
	matched := g.engine.evaluate(text, forInput)
	if matched == nil {
		return promptguard.GuardVerdict{
			GuardName: g.name,
			Verdict:   promptguard.VerdictSafe,
			Reason:    "not match any rule",
			Layer:     "rule",
			Metadata:  nil,
		}, nil
	}

	return promptguard.GuardVerdict{
		GuardName: g.name,
		Verdict:   promptguard.Verdict(matched.Verdict),
		Reason:    "matched rule " + matched.Name,
		Layer:     "rule",
		Metadata: map[string]any{
			"matched_rule": matched.Name,
			"tags":         matched.Tags,
		},
	}, nil
}

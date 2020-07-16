package rules

import "time"

// Rule represent a set of configuration for a break or a pomodori
type Rule struct {
	Rounds   int
	Duration time.Duration
}

// Rules wraps together pomodori and breaks rules
type Rules struct {
	Pomodori Rule
	Breaks   Rule
}

// NewTestRules inits fresh set of rules used for small tests
func NewTestRules() Rules {
	return Rules{Pomodori: Rule{Rounds: 4, Duration: 5 * time.Second}, Breaks: Rule{Rounds: 3, Duration: 5 * time.Second}}
}

// Package state handles the state of an entire pomodoro session
package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/papey/calabash/internal/rules"
)

// Status represents either a Working period or a Break period
type Status int

const (
	// Working status
	Working Status = 0
	// Break status
	Break Status = 1
)

// Print status to a human readable value
func (st Status) String() string {
	switch st {
	case 0:
		return "W"
	case 1:
		return "B"
	}

	return "Unknown"
}

// Session represent a pomodo instance state
type Session struct {
	// Time elapsed for current session (break or pomodori)
	Elapsed time.Duration
	// Time remaining for current session (break or pomodori)
	Remaining time.Duration
	// A pointer to when this pomodoro session started
	StartedAt time.Time
	// A pointer to when this pomodoro session finished
	FinishedAt time.Time
	// Is this session running (can be paused)
	Running bool
	// Is this session started
	Started bool
	// Is this session finished
	Finished bool
	// Is this current session a working one or a break one
	Status Status
	// Count the number of past pomodori for this entire session
	Pomodori int
	// Count the number of past breaks for this entire session
	Breaks int
	// Check if this beak is the last and long beak in this pomodoro session
	LongBreak bool
	// Set of rules for this pomodoro
	Rules rules.Rules
	// Lock used to avoid race conditions when updating and reading the session
	Lock sync.RWMutex `json:"-"`
	// A side channel used to cancel a session
	Cancel chan bool `json:"-"`
}

// Start a new pomodo session
func (s *Session) Start() {
	s.StartedAt = time.Now()
	s.Started = true
	s.Status = Working
}

// NewSession creates a new pomodoro session
func NewSession() Session {
	return Session{Cancel: make(chan bool, 1), Rules: rules.NewTestRules()}
}

func (s *Session) update() bool {
	// If this session is not running, just return
	if !s.Running {
		return false
	}

	// Lock
	s.Lock.Lock()
	// Unlock at the end of this func
	defer s.Lock.Unlock()

	// Update time elapsed
	s.Elapsed += 1 * time.Second

	// Check status : working time or break time ?
	if s.Status == Working {
		// Update remaining
		s.Remaining = s.Rules.Pomodori.Duration - s.Elapsed

		// If Working time is over
		if s.Elapsed >= s.Rules.Pomodori.Duration {
			// Increment pomodori counter
			s.Pomodori++

			// If it's the last pomodori for this session, long break
			if s.Pomodori == s.Rules.Pomodori.Rounds {
				fmt.Println("This session is over take a long break !")
				s.LongBreak = true
			} else {
				fmt.Println("Take a little break")
			}

			// Set status to break
			s.Status = Break
			// Init elapsed counter back to 0
			s.Elapsed = 0
		}
	} else {

		// Is this break a long break ?
		if s.LongBreak {
			// Update remaining
			s.Remaining = s.Rules.Pomodori.Duration*4 - s.Elapsed

			if s.Elapsed >= s.Rules.Breaks.Duration*4 {
				fmt.Println("Done")
				s.Terminate()
				// Everything is done
				return true
			}

		} else if s.Elapsed >= s.Rules.Breaks.Duration {
			// Update remaining
			s.Remaining = s.Rules.Pomodori.Duration - s.Elapsed

			// Increment break counter
			s.Breaks++
			fmt.Println("Get back to work")
			// Set status to working
			s.Status = Working
			// Init elapsed counter back to 0
			s.Elapsed = 0
			// Update remaining
			s.Remaining = s.Rules.Breaks.Duration - s.Elapsed
		}

	}

	// Do not stop this running loop
	return false
}

// Toogle a session between running and paused mode
func (s *Session) Toogle() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.Running = !s.Running
}

// Terminate sets a finished session to not running and finished
func (s *Session) Terminate() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.Running = false
	s.Started = false
	s.Finished = true
	s.FinishedAt = time.Now()
}

// Run starts a ticker within a session
func (s *Session)Run() {
	s.Running = true
	// Start a new ticker, tick on every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Wait for ticks
	for {
		select {
		case <-s.Cancel:
			fmt.Println("Session canceled")
			return
		case t := <-ticker.C:
			fmt.Println("Tick, tac, tick, tac", t)
			// Update state
			if stop := s.update(); stop {
				// If a stop is returned, exit this function
				return
			}
		}
	}
}

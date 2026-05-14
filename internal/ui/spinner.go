package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

type SpinnerWrapper struct {
	msg      string
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
}

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func NewSpinner(msg string) *SpinnerWrapper {
	return &SpinnerWrapper{
		msg:      msg,
		running:  false,
		stopChan: make(chan struct{}),
	}
}

func (s *SpinnerWrapper) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		idx := 0
		spinnerColor := color.New(color.FgCyan).SprintFunc()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s", spinnerColor(spinnerChars[idx%len(spinnerChars)]), s.msg)
				idx++
			}
		}
	}()
}

func (s *SpinnerWrapper) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	fmt.Printf("\r%s %s\n", SuccessColor("✓"), s.msg)
	s.stopChan = make(chan struct{})
}

func (s *SpinnerWrapper) StopWithError(err error) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	fmt.Printf("\r%s %s\n", ErrorColor("✗"), s.msg)
	Error(err.Error())
	s.stopChan = make(chan struct{})
}

func (s *SpinnerWrapper) Update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

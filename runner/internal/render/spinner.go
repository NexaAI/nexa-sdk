package render

import (
	"sync"

	"github.com/charmbracelet/huh/spinner"
)

type Spinner struct {
	*spinner.Spinner
	start sync.WaitGroup
	stop  sync.WaitGroup
}

func NewSpinner(desc string) *Spinner {
	s := Spinner{
		Spinner: spinner.New().Title(desc).Type(spinner.Globe),
	}

	s.Action(func() { s.start.Wait() })

	return &s
}

func (s *Spinner) Start() {
	s.start.Add(1)
	s.stop.Add(1)

	go func() {
		s.Spinner.Run()
		s.stop.Done()
	}()
}

func (s *Spinner) Stop() {
	s.start.Done()
	s.stop.Wait()
}

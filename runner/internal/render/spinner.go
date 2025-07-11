package render

import (
	"github.com/charmbracelet/huh/spinner"
)

type Spinner struct {
	*spinner.Spinner
	ch chan struct{}
}

func NewSpinner(desc string) *Spinner {
	spin := spinner.New().Title(desc).Type(spinner.Globe)

	ch := make(chan struct{})
	spin.Action(func() {
		<-ch
	})

	return &Spinner{
		Spinner: spin,
		ch:      ch,
	}
}

func (s *Spinner) Start() {
	go s.Spinner.Run()
}

func (s *Spinner) Stop() {
	s.ch <- struct{}{}
}

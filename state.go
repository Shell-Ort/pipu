package main

import (
	"fmt"
	"os"
)

// State Estructura para mover el cursor e imprimir por pantalla donde se quiera
type State struct {
	index int
}

func NewState() *State {
	return new(State)
}

func (s *State) Up(n int) {
	s.index -= n
	sal := fmt.Sprintf("\u001B[%dA", n)
	_, _ = os.Stdout.WriteString(sal)
}

func (s *State) Down(n int) {
	s.index += n
	sal := fmt.Sprintf("\033[%dB", n)
	_, _ = os.Stdout.WriteString(sal)
}

func (s *State) GoTo(line int) {
	if line > s.index {
		s.Down(line - s.index)
	} else if line < s.index {
		s.Up(s.index - line)
	}
}

func (s *State) Clear() {
	_, _ = os.Stdout.WriteString("\r\033[2K")
}

func (s *State) Write(text string) {
	s.Clear()
	_, _ = os.Stdout.WriteString(text)
}

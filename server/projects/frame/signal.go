// +build windows
package frame

import (
	"os"
)

type ISignalQuitDealer interface {
	Quit()
}

type SignalDealer struct {
	SignalChan     chan os.Signal
	quitDealer     ISignalQuitDealer
	CpuProfileFlag bool
	CpuFile        *os.File
}

func (s *SignalDealer) SetSignalQuitDealer(dealer ISignalQuitDealer) {
	s.quitDealer = dealer
}

var GSignalDealer *SignalDealer

func init() {
	GSignalDealer = &SignalDealer{
		SignalChan: make(chan os.Signal),
		quitDealer: nil,
	}
}

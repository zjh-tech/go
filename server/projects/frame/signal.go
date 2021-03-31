// +build windows
package frame

import (
	"os"
)

type ISignalQuitDealer interface {
	Quit()
}

type SignalDealer struct {
	signal_chan      chan os.Signal
	quit_dealer      ISignalQuitDealer
	cpu_profile_flag bool
	cpu_file         *os.File
}

func (s *SignalDealer) SetSignalQuitDealer(dealer ISignalQuitDealer) {
	s.quit_dealer = dealer
}

var GSignalDealer *SignalDealer

func init() {
	GSignalDealer = &SignalDealer{
		signal_chan: make(chan os.Signal),
		quit_dealer: nil,
	}
}

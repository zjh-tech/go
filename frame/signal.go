// +build windows
package frame

import (
	"os"
	"os/signal"
	"syscall"
)

type ISignalQuitDealer interface {
	Quit()
}

type SignalDealer struct {
	signal_chan chan os.Signal
	quit_dealer ISignalQuitDealer
}

func (s *SignalDealer) SetSignalQuitDealer(dealer ISignalQuitDealer) {
	s.quit_dealer = dealer
}

func (s *SignalDealer) RegisterSigHandler() {
	signal.Notify(s.signal_chan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func (s *SignalDealer) ListenSignal() {
	go func() {
		for {
			signal := <-s.signal_chan
			switch signal {
			case syscall.SIGINT:
				{
					ELog.Info("HANDLE SIGINT SIGNAL")
					if s.quit_dealer != nil {
						s.quit_dealer.Quit()
					}
				}
			case syscall.SIGTERM:
				{
					ELog.Info("HANDLE SIGTERM SIGNAL")
					if s.quit_dealer != nil {
						s.quit_dealer.Quit()
					}
				}
			case syscall.SIGQUIT:
				{
					ELog.Info("HANDLE SIGQUIT SIGNAL")
					if s.quit_dealer != nil {
						s.quit_dealer.Quit()
					}
				}
			default:
				{
					ELog.Error("Single=%v not Handler", signal)
				}
			}
		}
	}()
}

var GSignalDealer *SignalDealer

func init() {
	GSignalDealer = &SignalDealer{
		signal_chan: make(chan os.Signal),
		quit_dealer: nil,
	}
}

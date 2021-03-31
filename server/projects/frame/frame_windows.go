// +build windows
package frame

import (
	"os/signal"
	"syscall"

	"projects/go-engine/elog"
)

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
					elog.Info("HANDLE SIGINT SIGNAL")
					if s.quit_dealer != nil {
						s.quit_dealer.Quit()
					}
				}
			case syscall.SIGTERM:
				{
					elog.Info("HANDLE SIGTERM SIGNAL")
					if s.quit_dealer != nil {
						s.quit_dealer.Quit()
					}
				}
			case syscall.SIGQUIT:
				{
					elog.Info("HANDLE SIGQUIT SIGNAL")
					if s.quit_dealer != nil {
						s.quit_dealer.Quit()
					}
				}
			default:
				{
					elog.Error("Single=%v not Handler", signal)
				}
			}
		}
	}()
}

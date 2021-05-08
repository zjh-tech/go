// +build windows
package frame

import (
	"os/signal"
	"syscall"
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

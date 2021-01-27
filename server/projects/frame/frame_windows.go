// +build windows
package frame

import (
	"os/signal"
	"syscall"

	"projects/go-engine/elog"
)

func (s *SignalDealer) RegisterSigHandler() {
	signal.Notify(s.SignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func (s *SignalDealer) ListenSignal() {
	go func() {
		for {
			signal := <-s.SignalChan
			switch signal {
			case syscall.SIGINT:
				{
					elog.Info("HANDLE SIGINT SIGNAL")
					if s.quitDealer != nil {
						s.quitDealer.Quit()
					}
				}
			case syscall.SIGTERM:
				{
					elog.Info("HANDLE SIGTERM SIGNAL")
					if s.quitDealer != nil {
						s.quitDealer.Quit()
					}
				}
			case syscall.SIGQUIT:
				{
					elog.Info("HANDLE SIGQUIT SIGNAL")
					if s.quitDealer != nil {
						s.quitDealer.Quit()
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

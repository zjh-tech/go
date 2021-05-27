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
	signalChan chan os.Signal
	quitDealer ISignalQuitDealer
}

func (s *SignalDealer) Init(dealer ISignalQuitDealer) {
	s.quitDealer = dealer

	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for {
			signal := <-s.signalChan
			switch signal {
			case syscall.SIGINT:
				{
					ELog.Info("HANDLE SIGINT SIGNAL")
					if s.quitDealer != nil {
						s.quitDealer.Quit()
					}
				}
			case syscall.SIGTERM:
				{
					ELog.Info("HANDLE SIGTERM SIGNAL")
					if s.quitDealer != nil {
						s.quitDealer.Quit()
					}
				}
			case syscall.SIGQUIT:
				{
					ELog.Info("HANDLE SIGQUIT SIGNAL")
					if s.quitDealer != nil {
						s.quitDealer.Quit()
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
		signalChan: make(chan os.Signal),
		quitDealer: nil,
	}
}

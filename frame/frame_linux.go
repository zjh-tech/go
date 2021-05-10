// +build linux
package frame

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func (s *SignalDealer) RegisterSigHandler() {
	signal.Notify(s.signal_chan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
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
			case syscall.SIGUSR1:
				{
					if s.cpu_profile_flag == false {
						s.cpu_profile_flag = true
						ELog.Info("HANDLE SIGQUIT SIGUSR1 : Start Cpu Profile")
						s.cpu_file, _ = os.Create("cpu_profile_file")
						pprof.StartCPUProfile(s.cpu_file)
					} else {
						s.cpu_profile_flag = false
						ELog.Info("HANDLE SIGQUIT SIGUSR1 : Stop Cpu Profile")
						pprof.StopCPUProfile()
						s.cpu_file.Close()
						s.cpu_file = nil
					}
				}
			case syscall.SIGUSR2:
				{

				}
			default:
				{
					ELog.Error("Single=%v not Handler", signal)
				}
			}
		}
	}()
}

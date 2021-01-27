// +build linux
package frame

import (
	"os"
	"os/signal"
	"projects/go-engine/elog"
	"runtime/pprof"
	"syscall"
)

func (s *SignalDealer) RegisterSigHandler() {
	signal.Notify(s.SignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
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
			case syscall.SIGUSR1:
				{
					if s.CpuProfileFlag == false {
						s.CpuProfileFlag = true
						elog.Info("HANDLE SIGQUIT SIGUSR1 : Start Cpu Profile")
						s.CpuFile, _ = os.Create("cpu_profile_file")
						pprof.StartCPUProfile(s.CpuFile)
					} else {
						s.CpuProfileFlag = false
						elog.Info("HANDLE SIGQUIT SIGUSR1 : Stop Cpu Profile")
						pprof.StopCPUProfile()
						s.CpuFile.Close()
						s.CpuFile = nil
					}
				}
			case syscall.SIGUSR2:
				{

				}
			default:
				{
					elog.Error("Single=%v not Handler", signal)
				}
			}
		}
	}()
}

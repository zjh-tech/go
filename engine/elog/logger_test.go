package elog

import "testing"

func Benchmark_Info(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		TestLogger.Info("Benchmark Info")
	}
}

func Benchmark_InfoA(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		TestLogger.InfoA("Benchmark InfoA")
	}
}

var TestLogger *Logger

func init() {
	path := "./log"
	level := 1
	TestLogger = NewLogger(path, level)
	TestLogger.Init()
}

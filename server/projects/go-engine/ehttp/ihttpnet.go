package ehttp

type INet interface {
	PushHttpEvent(IHttpEvent)
	Run(loop_count int) bool
}

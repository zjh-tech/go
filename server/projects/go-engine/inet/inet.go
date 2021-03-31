package inet

type INet interface {
	PushEvent(IEvent)
	Connect(addr string, sess ISession)
	Listen(addr string, factory ISessionFactory, listen_max_count int) bool
	Run(loop_count int) bool
}

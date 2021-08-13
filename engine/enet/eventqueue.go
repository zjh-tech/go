package enet

//eventQueue
type EventQueue struct {
	evtQueue chan interface{}
}

func newEventQueue(maxCount uint32) *EventQueue {
	return &EventQueue{
		evtQueue: make(chan interface{}, maxCount),
	}
}

func (e *EventQueue) PushEvent(req interface{}) {
	e.evtQueue <- req
}

func (e *EventQueue) GetEventQueue() chan interface{} {
	return e.evtQueue
}

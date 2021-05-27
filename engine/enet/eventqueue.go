package enet

//eventQueue
type EventQueue struct {
	evtQueue chan interface{}
}

func new_event_queue(max_count uint32) *EventQueue {
	return &EventQueue{
		evtQueue: make(chan interface{}, max_count),
	}
}

func (e *EventQueue) PushEvent(req interface{}) {
	e.evtQueue <- req
}

func (e *EventQueue) GetEventQueue() chan interface{} {
	return e.evtQueue
}

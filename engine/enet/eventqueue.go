package enet

//eventQueue
type EventQueue struct {
	evt_queue chan interface{}
}

func new_event_queue(max_count uint32) *EventQueue {
	return &EventQueue{
		evt_queue: make(chan interface{}, max_count),
	}
}

func (e *EventQueue) PushEvent(req interface{}) {
	e.evt_queue <- req
}

func (e *EventQueue) GetEventQueue() chan interface{} {
	return e.evt_queue
}

package enet

import (
	"projects/go-engine/inet"
)

//Event
type Event struct {
	event_type uint32
	conn       inet.IConnection
	datas      interface{}
}

func NewEvent(t uint32, c inet.IConnection, datas interface{}) *Event {
	return &Event{
		event_type: t,
		conn:       c,
		datas:      datas,
	}
}

func (e *Event) GetType() uint32 {
	return e.event_type
}

func (e *Event) GetConn() inet.IConnection {
	return e.conn
}

func (e *Event) GetDatas() interface{} {
	return e.datas
}

type EventQueue struct {
	evt_queue chan inet.IEvent
}

func new_event_queue(max_count uint32) *EventQueue {
	return &EventQueue{
		evt_queue: make(chan inet.IEvent, max_count),
	}
}

func (e *EventQueue) PushEvent(req inet.IEvent) {
	e.evt_queue <- req
}

func (e *EventQueue) GetEventQueue() chan inet.IEvent {
	return e.evt_queue
}

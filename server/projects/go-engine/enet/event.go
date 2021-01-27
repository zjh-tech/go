package enet

import (
	"projects/go-engine/inet"
)

//Event
type Event struct {
	eventType uint32
	conn      inet.IConnection
	datas     interface{}
}

func NewEvent(t uint32, c inet.IConnection, datas interface{}) *Event {
	return &Event{
		eventType: t,
		conn:      c,
		datas:     datas,
	}
}

func (e *Event) GetType() uint32 {
	return e.eventType
}

func (e *Event) GetConn() inet.IConnection {
	return e.conn
}

func (e *Event) GetDatas() interface{} {
	return e.datas
}

//EventQueue
type EventQueue struct {
	evtQueue chan inet.IEvent
}

func NewEventQueue(maxCount uint32) *EventQueue {
	return &EventQueue{
		evtQueue: make(chan inet.IEvent, maxCount),
	}
}

func (e *EventQueue) PushEvent(req inet.IEvent) {
	e.evtQueue <- req
}

func (e *EventQueue) GetEventQueue() chan inet.IEvent {
	return e.evtQueue
}

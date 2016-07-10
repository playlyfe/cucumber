package core

type EventType int

const (
	TestRunStarting  EventType = 0
	TestCaseStarting EventType = 1
	TestStepStarting EventType = 2
	TestStepFinished EventType = 3
	TestCaseFinished EventType = 4
	TestRunFinished  EventType = 5
)

type Event struct {
	Name EventType
	Data interface{}
}

type EventHandler func(event *Event)

type EventBus struct {
	handlers map[EventType][]EventHandler
}

func (e *EventBus) Broadcast(eventType EventType, data interface{}) {
	for _, handler := range e.handlers[eventType] {
		handler(&Event{
			Name: eventType,
			Data: data,
		})
	}
}

func (e *EventBus) RegisterHandler(eventType EventType, handler EventHandler) {
	handlers, ok := e.handlers[eventType]
	if !ok {
		handlers = []EventHandler{}
	}
	handlers = append(handlers, handler)
	e.handlers[eventType] = handlers
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: map[EventType][]EventHandler{},
	}
}

package commands

type Value interface{}

type EventName string

const (
	PrintStarted  EventName = "PrintStarted"
	PrintStopped  EventName = "PrintStopped"
	PrintPaused   EventName = "PrintPaused"
	PrintResumed  EventName = "PrintResumed"
	PrintFailed   EventName = "PrintFailed"
	PrintError    EventName = "PrintError"
	PrintFinished EventName = "PrintFinished"
	PrintPercent  EventName = "PrintPercent"
	PrintTime     EventName = "PrintTime"
)

type RegisteredEvent struct {
	Name     EventName
	Callback func()
}

var RegisteredEvents = []RegisteredEvent{}

func CreateListener(eventName EventName, callback func()) {
	RegisteredEvents = append(RegisteredEvents, RegisteredEvent{
		Name:     eventName,
		Callback: callback,
	})
}

func Emit(eventName EventName) {
	for _, event := range RegisteredEvents {
		if event.Name == eventName {
			event.Callback()
		}
	}
}

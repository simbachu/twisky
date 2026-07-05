package command

// Dispatcher will route write intents (post, like, follow, etc.) to command handlers.
// Commands are intentionally stubbed until twisky needs authenticated writes.
type Dispatcher struct{}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

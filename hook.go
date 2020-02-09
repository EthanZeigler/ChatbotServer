package botserver

// tracks the used ID numbers of registered hooks so they cna be later removed
var idCounter int32

type hookContainer struct {
	id   int32
	hook Hook
}

// Hook is any taken action to an input. Used in adding actions in individual channels.
type Hook interface {
	// Action called by the channel when
	Action(callback Callback) bool
	// name for debugging purposes
	Name() string
}

type BasicHook struct {
	Handler   func(callback Callback) bool
	DebugName string
}

type ConditionalHook interface {
	// Action called by the channel when
	Action(callback Callback) bool
	// name for debugging purposes
	Name() string

	Condition(callback Callback) bool
}

func (hook *BasicHook) Action(callback Callback) bool {
	return hook.Handler(callback)
}

func (hook *BasicHook) Name() string {
	return hook.DebugName
}

// Wraps a hook with an id to allow removal later
func wrapHook(h Hook) (c hookContainer) {
	c.hook = h
	c.id = idCounter
	idCounter++
	return
}

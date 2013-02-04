package actions

const (
	SvcUnset   = iota // The status hasn't been checked yet
	SvcUnknown        // Status checked but either doesn't exist or don't understand results
	SvcStarted
	SvcStopped
	SvcStarting
	SvcStopping
)

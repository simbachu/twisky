package intent

// Intent is a user-facing request to read or change application state.
type Intent interface {
	intent()
}

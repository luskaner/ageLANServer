package common

const (
	ErrSuccess = iota
	ErrGeneral
	ErrSignal
	// ErrLast is only used as a marker to where to start, not a real error
	ErrLast
)

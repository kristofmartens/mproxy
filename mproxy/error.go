package mproxy

import "fmt"

// Error code definitions
const (
	ErrorNoError = iota
	ErrorInvalidConfig
	ErrorProxyInitError
	ErrorHttpError
	ErrorKeyError
)

type Error struct {
	Code int
	Msg  string
	Err  error
}

func (e *Error) Error() string { return fmt.Sprintf("%d\t %q: %s", e.Code, e.Msg, e.Err) }

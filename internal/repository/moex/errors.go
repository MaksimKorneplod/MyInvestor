package moex

import (
	"errors"
	"fmt"
)

// Ошибки внешнего источника. Вызывающий код различает их через errors.Is и errors.As.
// ErrRequest — это ошибка аргументов до запроса, не ответ ISS.
var (
	ErrNetwork = errors.New("moex: network")
	ErrStatus  = errors.New("moex: status")
	ErrParse   = errors.New("moex: invalid response")
	ErrRequest = errors.New("moex: bad request")
)

// StatusError — ISS ответил кодом не 200.
type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("moex: status %d", e.Code)
}

func (e *StatusError) Unwrap() error { return ErrStatus }

// ParseError — HTTP 200, но тела недостаточно, чтобы собрать строку.
type ParseError struct {
	Reason string
}

func (e *ParseError) Error() string {
	if e.Reason == "" {
		return ErrParse.Error()
	}
	return "moex: invalid response: " + e.Reason
}

func (e *ParseError) Unwrap() error { return ErrParse }

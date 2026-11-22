package quotes

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

// Ошибки сервиса. HTTP-слой разбирает только их и не зависит от репозитория.
var (
	ErrInvalidRequest = errors.New("invalid quotes request")
	ErrNotFound       = errors.New("security not found")
	ErrUnavailable    = errors.New("market data source is temporarily unavailable")
	ErrBadUpstream    = errors.New("upstream returned bad data")
	ErrTimeout        = errors.New("upstream timeout")
	ErrRateLimited    = errors.New("upstream rate limited")
)

// UpstreamStatusError — ISS ответил кодом, который не сведён к отдельной ошибке.
type UpstreamStatusError struct {
	Code int
}

func (e *UpstreamStatusError) Error() string {
	return fmt.Sprintf("upstream status %d", e.Code)
}

// mapSource переводит ошибку репозитория в ошибку сервиса.
// Отмена вызывающего остаётся context.Canceled: обработчик не должен
// превращать её в ответ об отказе источника.
func mapSource(err error) error {
	if err == nil {
		return nil
	}
	if isServiceError(err) {
		return err
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	var status *moex.StatusError
	switch {
	case errors.Is(err, moex.ErrRequest):
		return fmt.Errorf("%w: %s", ErrInvalidRequest, err.Error())
	case errors.Is(err, moex.ErrParse):
		return fmt.Errorf("%w", ErrBadUpstream)
	case errors.Is(err, moex.ErrNetwork):
		return fmt.Errorf("%w", ErrUnavailable)
	case errors.As(err, &status):
		switch status.Code {
		case http.StatusTooManyRequests:
			return ErrRateLimited
		case http.StatusNotFound:
			return ErrNotFound
		default:
			return &UpstreamStatusError{Code: status.Code}
		}
	default:
		return err
	}
}

func isServiceError(err error) bool {
	var status *UpstreamStatusError
	return errors.As(err, &status) ||
		errors.Is(err, ErrInvalidRequest) ||
		errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrUnavailable) ||
		errors.Is(err, ErrBadUpstream) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrRateLimited)
}

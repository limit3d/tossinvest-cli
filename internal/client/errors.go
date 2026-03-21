package client

import (
	"errors"
	"fmt"
)

var ErrNoSession = errors.New("no active session")

type StatusError struct {
	StatusCode int
	Endpoint   string
	Body       string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status %d for %s", e.StatusCode, e.Endpoint)
}



type AuthError struct {
	StatusCode int
	Endpoint   string
	Body       string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authenticated request rejected with status %d for %s", e.StatusCode, e.Endpoint)
}

func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrNoSession) {
		return true
	}

	var authErr *AuthError
	return errors.As(err, &authErr)
}

func newStatusError(statusCode int, endpoint string, body []byte) error {
	bodyText := string(body)
	if statusCode == 401 || statusCode == 403 {
		return &AuthError{
			StatusCode: statusCode,
			Endpoint:   endpoint,
			Body:       bodyText,
		}
	}

	return &StatusError{
		StatusCode: statusCode,
		Endpoint:   endpoint,
		Body:       bodyText,
	}
}

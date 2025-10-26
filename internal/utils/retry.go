package utils

import (
	"errors"
	"time"
)

type RetriableError interface {
	error
	IsRetriable() bool
}

func WithRetry(fn func() error, attempts int, maxAttempts int) error {
	if maxAttempts == 0 {
		return fn()
	}
	if attempts >= maxAttempts {
		return errors.New("max attempts reached")
	}
	secondsToSleep := time.Duration(2*(attempts)+1) * time.Second
	var retriableErr RetriableError
	if err := fn(); err != nil {
		if errors.As(err, &retriableErr) && retriableErr.IsRetriable() {
			time.Sleep(secondsToSleep)
			return WithRetry(fn, attempts+1, maxAttempts)
		}
		return err
	}
	return nil
}

package model

import "fmt"

type ErrNotFound struct {
	Message string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}

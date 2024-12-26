package ws

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Name    string          `json:"name"`
	Code    string          `json:"code"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (err Error) Error() string {
	return err.Message
}

func NewActionNotFoundErr(cmd string) *Error {
	return &Error{
		Message: fmt.Sprintf("action not found: %s", cmd),
		Code:    "action_not_found",
	}
}

var (
	ErrChannelNotFound = Error{
		Code:    "channel_not_found",
		Message: "channel not found",
	}
)

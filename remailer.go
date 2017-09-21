package remailer

import (
	"errors"
)

type remailer struct {
	Dir string `json:"remailer_dir"`
}

// ErrReject indicates that this is a rejection message
var ErrReject = errors.New("rejected")

// Reject error
type Reject struct {
	Message string
	error
}

package remailer

import (
	"errors"

	"github.com/flashmob/go-guerrilla/response"
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

// BadRecipient is an error 550 without a comment
var BadRecipient = (&response.Response{
	EnhancedCode: response.BadDestinationMailboxAddress,
	BasicCode:    550,
	Class:        response.ClassPermanentFailure,
	Comment:      "",
}).String()

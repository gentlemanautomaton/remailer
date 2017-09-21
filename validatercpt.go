package remailer

import (
	"errors"
	"fmt"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) validateRCPT(e *mail.Envelope) (backends.Result, error) {
	rcptListSize := len(e.RcptTo)
	if rcptListSize == 0 {
		// not sure what we would do here, so we'll just punt.
		return nil, nil
	}
	rcpt := e.RcptTo[rcptListSize-1]
	addrs, kind, err := r.expandAddr(rcpt)
	if err != nil {
		if err.Error() == ErrReject.Error() {
			reject := err.(Reject).Message
			backends.Log().WithError(backends.NoSuchUser).Info("reject: " + err.(Reject).Message)
			return backends.NewResult(response.Canned.FailRcptCmd), errors.New(reject)
		}
		backends.Log().WithError(backends.NoSuchUser).Info("error: " + err.Error())
		return backends.NewResult(response.Canned.FailRcptCmd), backends.NoSuchUser
	}
	if len(addrs) == 0 {
		backends.Log().WithError(backends.NoSuchUser).Info("user not configured: " + rcpt.String())
		return backends.NewResult(response.Canned.FailRcptCmd), backends.NoSuchUser
	}
	// success?
	backends.Log().Info(fmt.Printf("OK: %s: %+v\n", kind, addrs))
	return backends.NewResult(response.Canned.SuccessRcptCmd), nil
}

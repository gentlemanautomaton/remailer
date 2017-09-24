package remailer

import (
	"errors"
	"fmt"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) saveMail(e *mail.Envelope) (backends.Result, error) {
	rcptListSize := len(e.RcptTo)
	if rcptListSize == 0 {
		// not sure what we would do here, so we'll just punt.
		backends.Log().WithError(backends.NoSuchUser).Info("addresses were supplied")
		return backends.NewResult(BadRecipient), backends.NoSuchUser
	}
	for _, rcpt := range e.RcptTo {
		addrs, kind, err := r.expandAddr(rcpt)
		if err != nil {
			if reject, ok := err.(Reject); ok {
				rejectMsg := reject.Message
				backends.Log().WithError(backends.NoSuchUser).Info("reject: " + rejectMsg)
				return backends.NewResult(BadRecipient), errors.New(rejectMsg)
			}
			backends.Log().WithError(backends.NoSuchUser).Info("error: " + err.Error())
			return backends.NewResult(BadRecipient), backends.NoSuchUser
		}
		if len(addrs) == 0 {
			backends.Log().WithError(backends.NoSuchUser).Info("user not configured: " + rcpt.String())
			return backends.NewResult(BadRecipient), backends.NoSuchUser
		}
		backends.Log().Info(fmt.Printf("OK: %s: %+v\n", kind, addrs))

		for _, addr := range addrs {
			if !addr.IsEmpty() && !addr.Address.IsEmpty() {
				return r.sendMessage(addr.Address, e)
			} else if addr.URL != nil {
				return r.postMessage(*addr.URL, e)
			} else if addr.SMTP != nil {
				return r.mxMessage(*addr.SMTP, e)
			}
		}
	}

	// success?
	return backends.NewResult(response.Canned.SuccessMessageQueued), nil
}

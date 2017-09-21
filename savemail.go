package remailer

import (
	"errors"
	"fmt"
	"io"
	"net/smtp"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) saveMail(e *mail.Envelope) (backends.Result, error) {
	sc, err := smtp.Dial(r.ForwarderAddr)
	if err != nil {
		backends.Log().WithError(backends.StorageNotAvailable).Info("smtp: " + err.Error())
		return backends.NewResult(response.Canned.FailReadErrorDataCmd), errors.New("Temporary Server Error, try again shortly")
	}
	defer sc.Close()

	rcptListSize := len(e.RcptTo)
	if rcptListSize == 0 {
		// not sure what we would do here, so we'll just punt.
		return nil, nil
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
			// TODO: possibly support a host:port combo in the forwarding value to just kick the message to that server?
			sc.Hello(r.HeloName)
			sc.Mail(e.MailFrom.String())
			sc.Rcpt(addr.String())
			w, err := sc.Data()
			if err != nil {
				// TODO: what happen
			}
			io.Copy(w, &e.Data)
			err = w.Close()
			if err != nil {
				// TODO: what happen
			}
		}
	}

	// success?
	return backends.NewResult(response.Canned.SuccessMessageQueued), nil
}

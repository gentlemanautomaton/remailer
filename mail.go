package remailer

import (
	"errors"
	"io"
	"net/smtp"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) sendMessage(addr mail.Address, e *mail.Envelope) (backends.Result, error) {
	sc, err := smtp.Dial(r.ForwarderAddr)
	if err != nil {
		backends.Log().WithError(backends.StorageNotAvailable).Info("smtp: " + err.Error())
		return backends.NewResult(response.Canned.FailReadErrorDataCmd), errors.New("Temporary Server Error, try again shortly")
	}
	sc.Close()

	err = sc.Hello(r.HeloName)
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:24?")
	}
	err = sc.Mail(e.MailFrom.String()) // I wonder if this should be something else, such as from a domain we control?  FIXME?
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:29?")
	}
	err = sc.Rcpt(addr.String())
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:34?")
	}
	w, err := sc.Data()
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:39?")
	}
	if w != nil {
		io.Copy(w, &e.Data)
	}
	err = w.Close()
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:47?")
	}

	return nil, nil
}

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

	sc.Hello(r.HeloName)
	sc.Mail(e.MailFrom.String())
	sc.Rcpt(addr.String())
	w, err := sc.Data()
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:27?")
	}
	io.Copy(w, &e.Data)
	err = w.Close()
	if err != nil {
		// TODO: what happen
		backends.Log().WithError(err).Info("mail.go:33?")
	}

	return nil, nil
}

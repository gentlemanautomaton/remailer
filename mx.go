package remailer

import (
	"bytes"
	"errors"
	"io"
	"net/smtp"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) mxMessage(hp HostPort, e *mail.Envelope) (backends.Result, error) {
	sc, err := smtp.Dial(hp.String())
	if err != nil {
		backends.Log().WithError(backends.StorageNotAvailable).Info("smtp: " + err.Error())
		return backends.NewResult(response.Canned.FailReadErrorDataCmd), errors.New("Temporary Server Error, try again shortly")
	}

	sc.Hello(r.HeloName)
	sc.Mail(e.MailFrom.String())
	for _, addr := range e.RcptTo {
		sc.Rcpt(addr.String())
	}
	w, err := sc.Data()
	if err != nil {
		// TODO: what happen
	}
	io.Copy(w, bytes.NewBuffer(e.Data.Bytes()))
	err = w.Close()
	if err != nil {
		// TODO: what happen
	}

	return nil, nil
}

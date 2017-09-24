package remailer

import (
	"errors"
	"io"
	"net/smtp"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) sendMessage(smtpConn *smtp.Client, addr mail.Address, e *mail.Envelope) (backends.Result, error) {
	if smtpConn == nil {
		var err error
		smtpConn, err = smtp.Dial(r.ForwarderAddr)
		if err != nil {
			backends.Log().WithError(backends.StorageNotAvailable).Info("smtp: " + err.Error())
			return backends.NewResult(response.Canned.FailReadErrorDataCmd), errors.New("Temporary Server Error, try again shortly")
		}
	}

	smtpConn.Hello(r.HeloName)
	smtpConn.Mail(e.MailFrom.String())
	smtpConn.Rcpt(addr.String())
	w, err := smtpConn.Data()
	if err != nil {
		// TODO: what happen
	}
	io.Copy(w, &e.Data)
	err = w.Close()
	if err != nil {
		// TODO: what happen
	}

	return nil, nil
}

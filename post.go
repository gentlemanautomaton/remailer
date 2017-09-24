package remailer

import (
	"net/url"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
)

func (r *remailer) postMessage(u url.URL, e *mail.Envelope) (backends.Result, error) {
	//e.String()
	return nil, nil
}

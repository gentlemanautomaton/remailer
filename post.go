package remailer

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) postMessage(u url.URL, e *mail.Envelope) (backends.Result, error) {
	res, err := http.Post(u.String(), "text/plain", e.NewReader())
	if err != nil {
		backends.Log().WithError(backends.StorageNotAvailable).Info("post: " + err.Error())
		return backends.NewResult(response.Canned.FailReadErrorDataCmd), errors.New("Temporary Server Error, try again shortly")
	}
	backends.Log().Info("post: " + res.Status)

	return nil, nil // FIXME
}

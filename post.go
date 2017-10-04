package remailer

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) postMessage(u url.URL, e *mail.Envelope, dBuf *bytes.Buffer) (backends.Result, error) {
	res, err := http.Post(u.String(), "text/plain", io.MultiReader(
		strings.NewReader(e.DeliveryHeader),
		bytes.NewReader(dBuf.Bytes()),
	))
	if err != nil {
		backends.Log().WithError(backends.StorageNotAvailable).Info("post: " + err.Error())
		return backends.NewResult(response.Canned.FailReadErrorDataCmd), errors.New("Temporary Server Error, try again shortly")
	}
	backends.Log().Info("post: " + res.Status)

	return nil, nil // FIXME
}

package remailer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func (r *remailer) saveMail(e *mail.Envelope) (backends.Result, error) {
	e.Lock()
	defer e.Unlock()
	rcptListSize := len(e.RcptTo)
	if rcptListSize == 0 {
		// not sure what we would do here, so we'll just punt.
		backends.Log().WithError(backends.NoSuchUser).Info("no addresses were supplied")
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

		dBuf := make([]io.Writer, len(addrs))
		for i := range dBuf {
			dBuf[i] = new(bytes.Buffer)
		}
		w := io.MultiWriter(dBuf...)
		io.Copy(w, &e.Data)

		for i, addr := range addrs {
			if !addr.IsEmpty() && !addr.Address.IsEmpty() {
				if be, err := r.sendMessage(addr.Address, e, dBuf[i].(*bytes.Buffer)); err != nil {
					return be, err
				}
			} else if addr.URL != nil {
				if be, err := r.postMessage(*addr.URL, e, dBuf[i].(*bytes.Buffer)); err != nil {
					return be, err
				}
			} else if addr.SMTP != nil {
				if be, err := r.mxMessage(*addr.SMTP, e, dBuf[i].(*bytes.Buffer)); err != nil {
					return be, err
				}
			}
		}
	}

	// success?
	return backends.NewResult(response.Canned.SuccessMessageQueued), nil
}

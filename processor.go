package remailer

import (
	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

// Processor is a Guerrila backend processor to provide email forwarding service
func Processor() backends.Decorator {
	response.Canned.FailRcptCmd = (&response.Response{
		EnhancedCode: response.BadDestinationMailboxAddress,
		BasicCode:    550,
		Class:        response.ClassPermanentFailure,
		Comment:      "",
	}).String()

	var r *remailer

	backends.Svc.AddInitializer(backends.InitializeWith(func(configData backends.BackendConfig) error {
		configType := backends.BaseConfig(&remailer{})
		bc, err := backends.Svc.ExtractConfig(configData, configType)
		if err != nil {
			return err
		}

		r = bc.(*remailer)

		backends.Log().Info("FRED: " + r.Dir)

		return nil
	}))

	return func(p backends.Processor) backends.Processor {
		return backends.ProcessWith(func(e *mail.Envelope, task backends.SelectTask) (backends.Result, error) {
			switch task {
			case backends.TaskValidateRcpt:
				return r.validateRCPT(e)
			case backends.TaskSaveMail:
				return r.saveMail(e)
			default:
				return nil, nil
			}
		})
	}
}

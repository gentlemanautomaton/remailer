package remailer

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

type remailer struct {
	Dir string `json:"remailer_dir"`
}

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
			if task == backends.TaskValidateRcpt {
				if size := len(e.RcptTo); size > 0 {
					rcpt := e.RcptTo[size-1]
					addrs, kind, err := r.expandAddr(rcpt)
					if err != nil {
						if err.Error() == ErrReject.Error() {
							reject := err.(Reject).Message
							backends.Log().WithError(backends.NoSuchUser).Info("reject: " + err.(Reject).Message)
							return backends.NewResult(response.Canned.FailRcptCmd), errors.New(reject)
						}
						backends.Log().WithError(backends.NoSuchUser).Info("error: " + err.Error())
						return backends.NewResult(response.Canned.FailRcptCmd), backends.NoSuchUser
					}
					if len(addrs) == 0 {
						backends.Log().WithError(backends.NoSuchUser).Info("user not configured: " + rcpt.String())
						return backends.NewResult(response.Canned.FailRcptCmd), backends.NoSuchUser
					}
					// success?
					backends.Log().Info(fmt.Printf("OK: %s: %+v\n", kind, addrs))
					return backends.NewResult(response.Canned.SuccessRcptCmd), nil
				}
			}
			if task == backends.TaskSaveMail {
				if size := len(e.RcptTo); size > 0 {
					rcpt := e.RcptTo[size-1]
					addrs, kind, err := r.expandAddr(rcpt)
					if err != nil {
						if err.Error() == ErrReject.Error() {
							reject := err.(Reject).Message
							backends.Log().WithError(backends.NoSuchUser).Info("reject: " + err.(Reject).Message)
							return backends.NewResult(response.Canned.FailRcptCmd), errors.New(reject)
						}
						backends.Log().WithError(backends.NoSuchUser).Info("error: " + err.Error())
						return backends.NewResult(response.Canned.FailRcptCmd), backends.NoSuchUser
					}
					if len(addrs) == 0 {
						backends.Log().WithError(backends.NoSuchUser).Info("user not configured: " + rcpt.String())
						return backends.NewResult(response.Canned.FailRcptCmd), backends.NoSuchUser
					}
					// success?
					backends.Log().Info(fmt.Printf("OK: %s: %+v\n", kind, addrs))
					return backends.NewResult(response.Canned.SuccessMessageQueued), nil
				}
			}
			return nil, nil
		})
	}
}

// Address Kinds
const (
	KindUnknown        = ""
	KindNormal         = "normal"
	KindDomainWildcard = "domain-wildcard"
	KindPlus           = "plus"
	KindPlusFallback   = "plus-fallback"
)

func (r *remailer) expandAddr(rcpt mail.Address) (addrs []mail.Address, kind string, err error) {
	domainFilename := path.Join(r.Dir, strings.ToLower(rcpt.Host))

	// domain check
	var domain os.FileInfo
	domain, err = os.Stat(domainFilename)
	if err != nil {
		return nil, KindUnknown, err
	}

	// if this is a file, it's a wildcard or otherwise domain-wide config
	if !domain.IsDir() {
		addrs, err = getAddrsFromFile(domainFilename)
		kind = KindDomainWildcard
		if err != nil {
			return nil, KindUnknown, err
		}
		return
	}

	// user check
	// note: LHS of email addresses are supposed to be case-sensitive-able, but
	//   I'm making them case-insensitive deliberately as it's probably the right
	//   choice for nearly all situations...  (I'm also assuming + is special)
	userPieces := strings.SplitN(strings.ToLower(rcpt.User), "+", 2)
	userName := userPieces[0]
	userPlus := ""
	if len(userPieces) > 1 {
		userPlus = userPieces[1]
	}
	userNameFilename := path.Join(domainFilename, userName)
	var user os.FileInfo
	user, err = os.Stat(userNameFilename)
	if err != nil {
		return nil, KindUnknown, err
	}

	// if this is a file, it's a normal ol' user-wide config (meaning that + isn't handled specially)
	if !user.IsDir() {
		addrs, err = getAddrsFromFile(userNameFilename)
		kind = KindNormal
		if err != nil {
			return nil, KindUnknown, err
		}
		return
	}

	// userplus check
	userPlusFilename := path.Join(userNameFilename, userPlus)
	var plus os.FileInfo
	plus, err = os.Stat(userPlusFilename)
	kind = KindPlus
	if err != nil {
		userPlusFilename = path.Join(userNameFilename, "@") // fallback
		plus, err = os.Stat(userPlusFilename)
		kind = KindPlusFallback
		if err != nil {
			return nil, KindUnknown, err
		}
	}

	// if this is a file, it's a config for user+plus@domain.tld, or it's the user+@@domain.tld override config
	if !plus.IsDir() {
		addrs, err = getAddrsFromFile(userPlusFilename)
		// kind is already defined above (KindPlus|KindPlusFallback)
		if err != nil {
			return nil, KindUnknown, err
		}
		return
	}

	// if we got here, somethins is weird so we're ditching it.  (it seems to be a user+plus that is a directory, which isn't a thing... yet?)
	return nil, KindUnknown, err
}

// ErrReject indicates that this is a rejection message
var ErrReject = errors.New("rejected")

// Reject error
type Reject struct {
	Message string
	error
}

func getAddrsFromFile(filename string) ([]mail.Address, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var addrs = make([]mail.Address, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "reject:") {
			reject := strings.TrimPrefix(line, "reject:")
			return nil, Reject{reject, ErrReject}
		}
		addr, err := mail.NewAddress(line)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return addrs, nil
}

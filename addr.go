package remailer

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/flashmob/go-guerrilla/mail"
)

// Address Kinds
const (
	KindUnknown        = ""
	KindNormal         = "normal"
	KindDomainWildcard = "domain-wildcard"
	KindPlus           = "plus"
	KindPlusFallback   = "plus-fallback"
)

// Address is a superset of mail.Address plus url.URL
type Address struct {
	mail.Address
	URL  *url.URL
	SMTP *HostPort
}

// IsEmpty returns true if empty
func (a *Address) IsEmpty() bool {
	return a.URL == nil && a.SMTP == nil && a.Address.IsEmpty()
}

func (a *Address) String() string {
	if a.URL != nil {
		return a.URL.String()
	}
	if a.SMTP != nil {
		return a.SMTP.String()
	}
	if !a.Address.IsEmpty() {
		return a.Address.String()
	}
	return ""
}

// HostPort is a "hostname:1234"
type HostPort struct {
	Host string
	Port int
}

// ParseHostPort will return a HostPort from a string
func ParseHostPort(in string) (hp HostPort, err error) {
	host, portStr, err := net.SplitHostPort(in)
	if err != nil {
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return
	}
	return HostPort{host, port}, nil
}

func (h *HostPort) String() string {
	if h == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

func getAddrsFromFile(filename string) ([]Address, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var addrs = make([]Address, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "reject:") {
			reject := strings.TrimPrefix(line, "reject:")
			return nil, Reject{reject, ErrReject}
		}

		var addr Address

		if strings.HasPrefix(line, "https://") {
			// if HTTPS URL (we don't support non-HTTPS)
			u, err := url.Parse(line)
			if err != nil {
				return nil, err
			}
			addr = Address{URL: u}
		} else if strings.Contains(line, "://") {
			// we don't support this URL, so we are skipping this address...  FIXME?
			continue
		} else if strings.Contains(line, "@") {
			// if email address
			mAddr, err := mail.NewAddress(line)
			if err != nil {
				return nil, err
			}
			addr = Address{Address: mAddr}
		} else if strings.Contains(line, ":") {
			// if SMTP server address
			hp, err := ParseHostPort(line)
			if err != nil {
				return nil, err
			}
			addr = Address{SMTP: &hp}
		} else {
			// nothing matched, so we are skipping this address...  FIXME?
			continue
		}

		addrs = append(addrs, addr)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return addrs, nil
}

func (r *remailer) expandAddr(rcpt mail.Address) (addrs []Address, kind string, err error) {
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

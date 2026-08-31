package core

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Subdomain rules.
const (
	// MinSubdomainLen keeps generated and requested names from colliding with
	// single-letter operational records.
	MinSubdomainLen = 3
	// MaxSubdomainLen is the DNS label limit (RFC 1035 §2.3.4).
	MaxSubdomainLen = 63
	// GeneratedSubdomainLen is the length of an ephemeral name. With a
	// 32-character alphabet, 10 characters is ~50 bits — enough that a public
	// URL is not guessable by scanning.
	GeneratedSubdomainLen = 10
)

// generatedAlphabet omits vowels (so a random name cannot spell a word),
// and omits 0/1/l/o so a name read aloud or copied from a terminal is
// unambiguous.
const generatedAlphabet = "23456789bcdfghjkmnpqrstvwxyz"

var (
	ErrSubdomainTooShort  = errors.New("core: subdomain is too short")
	ErrSubdomainTooLong   = errors.New("core: subdomain exceeds the DNS label limit")
	ErrSubdomainCharset   = errors.New("core: subdomain may contain only lowercase letters, digits and hyphens")
	ErrSubdomainHyphen    = errors.New("core: subdomain may not start or end with a hyphen")
	ErrSubdomainReserved  = errors.New("core: subdomain is reserved")
	ErrSubdomainExhausted = errors.New("core: could not allocate a free subdomain")
)

// reservedSubdomains are names the relay must never hand to a tenant. Three
// groups, and each one is a real hazard rather than a nicety:
//
//   - operational hostnames the project itself serves or will serve, so a
//     tenant cannot occupy them later;
//   - names that impersonate the operator to a visitor ("admin", "login",
//     "billing") — a phishing page on a look-alike host is the abuse this
//     product attracts, per the candidate's care register;
//   - protocol and certificate-validation names ("_acme-challenge", "mx"),
//     where a tenant answering on the label could interfere with issuance.
var reservedSubdomains = map[string]bool{
	"_acme-challenge": true, "abuse": true, "account": true, "accounts": true,
	"admin": true, "administrator": true, "api": true, "app": true, "assets": true,
	"auth": true, "autoconfig": true, "autodiscover": true, "billing": true,
	"blog": true, "cdn": true, "checkout": true, "cname": true, "console": true,
	"dashboard": true, "dev": true, "developer": true, "devices": true, "dns": true,
	"docs": true, "download": true, "downloads": true, "email": true, "ftp": true,
	"git": true, "help": true, "imap": true, "internal": true, "localhost": true,
	"login": true, "mail": true, "manage": true, "mx": true, "ns": true, "ns1": true,
	"ns2": true, "pay": true, "payment": true, "payments": true, "pop": true,
	"portal": true, "pumasi": true, "register": true, "relay": true, "root": true,
	"secure": true, "security": true, "signin": true, "signup": true, "smtp": true,
	"ssh": true, "ssl": true, "staging": true, "static": true, "status": true,
	"support": true, "test": true, "tunnel": true, "user": true, "users": true,
	"verify": true, "webmail": true, "wildcard": true, "www": true,
}

// IsReservedSubdomain reports whether a name is on the reserved list.
func IsReservedSubdomain(name string) bool {
	return reservedSubdomains[strings.ToLower(strings.TrimSpace(name))]
}

// ValidateSubdomain checks a user-requested name against the DNS label rules
// and the reserved list. It is deliberately stricter than DNS: only lowercase
// letters, digits and inner hyphens, so a name cannot vary by case or by
// Unicode confusables between the certificate, the routing table and the
// address bar.
func ValidateSubdomain(name string) error {
	if len(name) < MinSubdomainLen {
		return fmt.Errorf("%w: %q is %d characters, minimum is %d", ErrSubdomainTooShort, name, len(name), MinSubdomainLen)
	}
	if len(name) > MaxSubdomainLen {
		return fmt.Errorf("%w: %q is %d characters, maximum is %d", ErrSubdomainTooLong, name, len(name), MaxSubdomainLen)
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-':
		default:
			return fmt.Errorf("%w: %q contains %q", ErrSubdomainCharset, name, string(c))
		}
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("%w: %q", ErrSubdomainHyphen, name)
	}
	if IsReservedSubdomain(name) {
		return fmt.Errorf("%w: %q", ErrSubdomainReserved, name)
	}
	return nil
}

// NameChecker reports whether a candidate name is already in use. *Registry
// satisfies it via Has.
type NameChecker interface {
	Has(subdomain string) bool
}

// GenerateSubdomain returns a random ephemeral name. The reader supplies the
// randomness so tests can pin it; pass nil for crypto/rand.
func GenerateSubdomain(randomness RandSource) (string, error) {
	if randomness == nil {
		randomness = cryptoRand{}
	}
	out := make([]byte, GeneratedSubdomainLen)
	for i := range out {
		n, err := randomness.Intn(len(generatedAlphabet))
		if err != nil {
			return "", fmt.Errorf("core: subdomain randomness: %w", err)
		}
		out[i] = generatedAlphabet[n]
	}
	return string(out), nil
}

// AllocateSubdomain generates a free ephemeral name, retrying on the
// vanishingly rare collision. It gives up after a bounded number of attempts
// rather than looping forever if the namespace is somehow saturated.
func AllocateSubdomain(taken NameChecker, randomness RandSource) (string, error) {
	const attempts = 8
	for i := 0; i < attempts; i++ {
		name, err := GenerateSubdomain(randomness)
		if err != nil {
			return "", err
		}
		// A generated name cannot be reserved (the alphabet has no vowels and
		// every reserved word does), but check anyway: the alphabet is a
		// constant someone may widen later.
		if IsReservedSubdomain(name) {
			continue
		}
		if taken != nil && taken.Has(name) {
			continue
		}
		return name, nil
	}
	return "", fmt.Errorf("%w after %d attempts", ErrSubdomainExhausted, attempts)
}

// RandSource yields uniform integers in [0,n). It exists so allocation is
// testable with a deterministic sequence.
type RandSource interface {
	Intn(n int) (int, error)
}

type cryptoRand struct{}

func (cryptoRand) Intn(n int) (int, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type GenerateOpts struct {
	Issuer      string
	AccountName string
}

type Key struct {
	secret      string
	issuer      string
	accountName string
}

func (k Key) Secret() string { return k.secret }

func (k Key) URL() string {
	issuerEscaped := url.QueryEscape(k.issuer)
	accountEscaped := url.QueryEscape(k.accountName)
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30", issuerEscaped, accountEscaped, k.secret, issuerEscaped)
}

// Generate creates a new secret and provisioning URI compatible with Google Authenticator.
func Generate(opts GenerateOpts) (Key, error) {
	issuer := strings.TrimSpace(opts.Issuer)
	if issuer == "" {
		issuer = "xbackend"
	}
	account := strings.TrimSpace(opts.AccountName)
	if account == "" {
		return Key{}, fmt.Errorf("account name required")
	}

	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return Key{}, err
	}

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	secret := encoder.EncodeToString(secretBytes)

	return Key{secret: secret, issuer: issuer, accountName: account}, nil
}

// Validate checks a 6-digit TOTP code for the current 30-second window with ±1 step skew.
func Validate(code, secret string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	now := time.Now()
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return false
	}

	timestep := int64(30)
	counter := now.Unix() / timestep
	for _, offset := range []int64{0, -1, 1} {
		if computeCode(secretBytes, counter+offset) == code {
			return true
		}
	}
	return false
}

func computeCode(secret []byte, counter int64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, secret)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	otp := truncated % 1000000
	return fmt.Sprintf("%06d", otp)
}

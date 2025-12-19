package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestValidatorParse(t *testing.T) {
	botToken := "test_bot_token"
	userJSON, _ := json.Marshal(telegramUser{ID: 42, FirstName: "John", LastName: "Doe", Username: "johnd"})

	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("user", string(userJSON))
	addHash(t, botToken, values)

	v := validator{botToken: botToken, initDataTTL: time.Hour}
	parsed, err := v.parse(values.Encode())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if parsed.ID != 42 || parsed.Username != "johnd" {
		t.Fatalf("unexpected parsed user: %+v", parsed)
	}
}

func TestValidatorParse_InvalidHash(t *testing.T) {
	botToken := "token"
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("user", `{"id":1}`)
	values.Set("hash", "invalid")

	v := validator{botToken: botToken, initDataTTL: time.Hour}
	if _, err := v.parse(values.Encode()); err == nil {
		t.Fatalf("expected error for invalid hash")
	}
}

func TestDisplayNameFallback(t *testing.T) {
	uc := UseCase{}
	dn, err := uc.displayName(telegramUser{ID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dn.String() != "tg_7" {
		t.Fatalf("unexpected display name: %s", dn.String())
	}
}

func TestNew_MissingBotToken(t *testing.T) {
	_, err := New(nil, nil, nil, nil, "", 0, 0, 0)
	if err == nil {
		t.Fatalf("expected error when bot token is missing")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_BOT_TOKEN") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func addHash(t *testing.T, botToken string, values url.Values) {
	t.Helper()

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var pairs []string
	for _, key := range keys {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	dataCheckString := strings.Join(pairs, "\n")

	secretHasher := hmac.New(sha256.New, []byte("WebAppData"))
	secretHasher.Write([]byte(botToken))
	secret := secretHasher.Sum(nil)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(h.Sum(nil)))
}

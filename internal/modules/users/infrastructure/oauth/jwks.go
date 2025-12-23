package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// remoteKeySet fetches and caches RSA public keys from a JWKS endpoint.
type remoteKeySet struct {
	url       string
	ttl       time.Duration
	client    *http.Client
	mu        sync.RWMutex
	fetchedAt time.Time
	keys      map[string]*rsa.PublicKey
}

func newRemoteKeySet(url string, ttl time.Duration) *remoteKeySet {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &remoteKeySet{
		url:    url,
		ttl:    ttl,
		client: &http.Client{Timeout: 5 * time.Second},
		keys:   map[string]*rsa.PublicKey{},
	}
}

func (s *remoteKeySet) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, errors.New("kid header is required")
	}
	now := time.Now()

	s.mu.RLock()
	cached := s.keys[kid]
	expired := s.fetchedAt.IsZero() || now.Sub(s.fetchedAt) > s.ttl
	s.mu.RUnlock()

	if cached != nil && !expired {
		return cached, nil
	}

	if err := s.refresh(ctx); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if k := s.keys[kid]; k != nil {
		return k, nil
	}
	return nil, fmt.Errorf("key %s not found", kid)
}

func (s *remoteKeySet) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var payload struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, jwk := range payload.Keys {
		if jwk.Kty != "RSA" || jwk.N == "" || jwk.E == "" || jwk.Kid == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			continue
		}
		var eInt int
		for _, b := range eBytes {
			eInt = eInt<<8 + int(b)
		}
		keys[jwk.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: eInt}
	}

	if len(keys) == 0 {
		return errors.New("no RSA keys found in JWKS")
	}

	s.mu.Lock()
	s.keys = keys
	s.fetchedAt = time.Now()
	s.mu.Unlock()
	return nil
}

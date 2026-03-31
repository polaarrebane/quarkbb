package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// TokenStatus is a status of auth token.
type TokenStatus string

// TokenStatus values.
const (
	RevokedTokenStatus   TokenStatus = "revoked"
	ValidTokenStatus     TokenStatus = "valid"
	UndefinedTokenStatus TokenStatus = "undefined"
)

// RevokedTokensStorage is a fast storage for revoked auth tokens.
type RevokedTokensStorage interface {
	Add(string) error
	Check(string) (TokenStatus, error)
}

type cacheRevokedStorage struct {
	mc *memcache.Client
}

// NewRevokedTokensStorage creates a new instance of RevokedTokensStorage.
func NewRevokedTokensStorage(mc *memcache.Client) RevokedTokensStorage {
	return &cacheRevokedStorage{
		mc: mc,
	}
}

func (rts cacheRevokedStorage) Add(jti string) error {
	now := time.Now().String()
	err := rts.mc.Set(&memcache.Item{
		Key: jti, Value: []byte(now), Expiration: 900,
	})
	if err != nil {
		return fmt.Errorf("add to revoked storage error: %w", err)
	}
	return nil
}

func (rts cacheRevokedStorage) Check(jti string) (TokenStatus, error) {
	_, err := rts.mc.Get(jti)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return ValidTokenStatus, nil
		}
		return UndefinedTokenStatus, fmt.Errorf("add to revoked storage error: %w", err)
	}
	return RevokedTokenStatus, nil
}

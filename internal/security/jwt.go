package security

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"codeberg.org/ronia/quarkbb/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService provides JWT token generation and verification functionality.
type JWTService interface {
	GenerateTokenPair(username string, userid string, sessionid string) (*TokenPair, error)
	VerifyAuthToken(tokenString string) (*AuthClaims, error)
	VerifyRefreshToken(tokenString string) (*RefreshClaims, error)
}

type jwtServiceImpl struct {
	privateKey      *ecdsa.PrivateKey
	publicKey       *ecdsa.PublicKey
	keyID           string
	audience        string
	issuer          string
	authTokenTTL    int
	refreshTokenTTL int
}

// NewJWTService creates a new JWTService instance with ECDSA key pair.
// It loads private and public keys from the configured secrets directory.
func NewJWTService(c config.Config) (JWTService, error) {
	keyID := c.Keys()

	path := "secrets/keys/" + keyID + "/private.pem"
	privateKey, err := loadPrivateKey(path)
	if err != nil {
		return nil, err
	}

	path = "secrets/keys/" + keyID + "/public.pem"
	publicKey, err := loadPublicKey(path)
	if err != nil {
		return nil, err
	}

	return &jwtServiceImpl{
		privateKey:      privateKey,
		publicKey:       publicKey,
		audience:        "https://quarkbb.org",
		issuer:          "https://quarkbb.org",
		authTokenTTL:    c.AuthTokenTTL(),
		refreshTokenTTL: c.RefreshTokenTTL(),
	}, nil
}

// TokenPair contains both access and refresh tokens with their expiration times.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresIn  int
	RefreshExpiresIn int
}

// AuthClaims represents the claims stored in JWT access tokens.
type AuthClaims struct {
	Username  string
	SessionID string
	jwt.RegisteredClaims
}

// RefreshClaims represents the claims stored in JWT refresh tokens.
type RefreshClaims struct {
	SessionID string
	jwt.RegisteredClaims
}

func (js *jwtServiceImpl) GenerateTokenPair(username string, userid string, sessionid string) (*TokenPair, error) {
	now := time.Now()

	return &TokenPair{
		AccessToken:      js.newAuthToken(username, userid, sessionid, now),
		RefreshToken:     js.newRefreshToken(userid, sessionid, now),
		AccessExpiresIn:  js.authTokenTTL,
		RefreshExpiresIn: js.refreshTokenTTL,
	}, nil
}

func (js *jwtServiceImpl) newAuthToken(username string, userid string, sessionid string, t time.Time) string {
	claims := &AuthClaims{
		Username:  username,
		SessionID: sessionid,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(time.Duration(js.authTokenTTL * int(time.Second)))),
			Issuer:    js.issuer,
			Subject:   userid,
			Audience:  jwt.ClaimStrings{js.audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = js.keyID
	r, _ := token.SignedString(js.privateKey)
	return r
}

func (js *jwtServiceImpl) newRefreshToken(userid string, sessionid string, t time.Time) string {
	claims := &RefreshClaims{
		SessionID: sessionid,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(time.Duration(js.refreshTokenTTL * int(time.Second)))),
			Issuer:    js.issuer,
			Subject:   userid,
			Audience:  jwt.ClaimStrings{js.audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = js.keyID
	r, _ := token.SignedString(js.privateKey)
	return r
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key file %q: %w", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an ECDSA private key")
	}
	return ecKey, nil
}

func loadPublicKey(path string) (*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file %q: %w", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	key, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	return key, nil
}

func (js *jwtServiceImpl) VerifyAuthToken(tokenString string) (*AuthClaims, error) {
	token, err := js.parseTokenString(tokenString, &AuthClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*AuthClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (js *jwtServiceImpl) VerifyRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := js.parseTokenString(tokenString, &RefreshClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (js *jwtServiceImpl) parseTokenString(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return js.publicKey, nil
		},
		jwt.WithAudience(js.audience),
		jwt.WithIssuer(js.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("verify auth token: %w", err)
	}
	return token, nil
}

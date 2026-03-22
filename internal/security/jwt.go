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

const authTokenTTL = 15        // minutes
const refreshTokenTTL = 24 * 7 // hours

// JWTService provides JWT token generation and verification functionality.
type JWTService interface {
	GenerateTokenPair(username string, userid string) (*TokenPair, error)
	VerifyAuthToken(tokenString string) (*AuthClaims, error)
	VerifyRefreshToken(tokenString string) (*RefreshClaims, error)
}

type jwtServiceImpl struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	keyID      string
	audience   string
	issuer     string
}

// New creates a new JWTService instance with ECDSA key pair.
// It loads private and public keys from the configured secrets directory.
func New(c config.Config) (JWTService, error) {
	keyID := c.GetKeys()

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
		privateKey: privateKey,
		publicKey:  publicKey,
		audience:   "https://quarkbb.org",
		issuer:     "https://quarkbb.org",
	}, nil
}

// TokenPair contains both access and refresh tokens with their expiration times.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresIn  time.Time
	RefreshExpiresIn time.Time
}

// AuthClaims represents the claims stored in JWT access tokens.
type AuthClaims struct {
	Username string
	jwt.RegisteredClaims
}

// RefreshClaims represents the claims stored in JWT refresh tokens.
type RefreshClaims struct {
	jwt.RegisteredClaims
}

func (js *jwtServiceImpl) GenerateTokenPair(username string, userid string) (*TokenPair, error) {
	now := time.Now()

	return &TokenPair{
		AccessToken:      js.newAuthToken(username, userid, now),
		RefreshToken:     js.newRefreshToken(userid, now),
		AccessExpiresIn:  now.Add(15 * time.Minute),
		RefreshExpiresIn: now.Add(refreshTokenTTL * time.Hour),
	}, nil
}

func (js *jwtServiceImpl) newAuthToken(username string, userid string, t time.Time) string {
	claims := &AuthClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(authTokenTTL * time.Minute)),
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

func (js *jwtServiceImpl) newRefreshToken(userid string, t time.Time) string {
	claims := &RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(refreshTokenTTL * time.Hour)),
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
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
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
	return jwt.ParseWithClaims(
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
}

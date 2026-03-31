// Package config loads and exposes application configuration from the config file.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds application configuration values loaded from config file.
type Config interface {
	AppKey() string
	Host() string
	Port() string
	Keys() string
	DSN() string
	Memcached() string
	AuthTokenTTL() int
	RefreshTokenTTL() int
}

type configImpl struct {
	appKey          string
	dsn             string
	host            string
	port            string
	keys            string
	memcached       string
	authTokenTTL    int
	refreshTokenTTL int
}

// New creates a new Config instance by reading configuration from file.
// It looks for 'main.json' in the './config/' directory and returns
// an error if the config file cannot be read or parsed.
func New() (Config, error) {
	viper.SetConfigName("main")
	viper.AddConfigPath("./config/")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	dsn := viper.GetString("dsn")
	host := viper.GetString("host")
	port := viper.GetString("port")
	keys := viper.GetString("keys")
	appKey := viper.GetString("app_key")
	authTokenTTL := viper.GetInt("auth_token_ttl")
	refreshTokenTTL := viper.GetInt("refresh_token_ttl")
	memcached := viper.GetString("memcached")

	return &configImpl{
		appKey:          appKey,
		dsn:             dsn,
		host:            host,
		port:            port,
		keys:            keys,
		authTokenTTL:    authTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		memcached:       memcached,
	}, nil
}

// AppKey returns the appl key from configuration.
func (c *configImpl) AppKey() string {
	return c.appKey
}

// Host returns the server host address from configuration.
func (c *configImpl) Host() string {
	return c.host
}

// Port returns the server port number from configuration.
func (c *configImpl) Port() string {
	return c.port
}

// Keys returns the key identifier for JWT token signing.
// This corresponds to the directory name containing the PEM key files.
func (c *configImpl) Keys() string {
	return c.keys
}

// DSN returns the database connection string.
// Used for establishing connection to the PostgreSQL database.
func (c *configImpl) DSN() string {
	return c.dsn
}

// Memcached returns connection string for memcached.
func (c *configImpl) Memcached() string {
	return c.memcached
}

// AuthTokenTTL returns the auth token ttl in seconds.
func (c *configImpl) AuthTokenTTL() int {
	return c.authTokenTTL
}

// RefreshTokenTTL returns the refresh token ttl in seconds.
func (c *configImpl) RefreshTokenTTL() int {
	return c.refreshTokenTTL
}

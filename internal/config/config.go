// Package config loads and exposes application configuration from the config file.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds application configuration values loaded from config file.
type Config struct {
	dsn             string
	host            string
	port            string
	keys            string
	authTokenTTL    int
	refreshTokenTTL int
}

// New creates a new Config instance by reading configuration from file.
// It looks for 'main.json' in the './config/' directory and returns
// an error if the config file cannot be read or parsed.
func New() (*Config, error) {
	viper.SetConfigName("main")
	viper.AddConfigPath("./config/")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	dsn := viper.GetString("dsn")
	host := viper.GetString("host")
	port := viper.GetString("port")
	keys := viper.GetString("keys")
	authTokenTTL := viper.GetInt("auth_token_ttl")
	refreshTokenTTL := viper.GetInt("refresh_token_ttl")

	return &Config{
		dsn:             dsn,
		host:            host,
		port:            port,
		keys:            keys,
		authTokenTTL:    authTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}, nil
}

// GetHost returns the server host address from configuration.
func (c *Config) GetHost() string {
	return c.host
}

// GetPort returns the server port number from configuration.
func (c *Config) GetPort() string {
	return c.port
}

// GetKeys returns the key identifier for JWT token signing.
// This corresponds to the directory name containing the PEM key files.
func (c *Config) GetKeys() string {
	return c.keys
}

// GetDSN returns the database connection string.
// Used for establishing connection to the PostgreSQL database.
func (c *Config) GetDSN() string {
	return c.dsn
}

// GetAuthTokenTTL returns the auth token ttl in seconds.
func (c *Config) GetAuthTokenTTL() int {
	return c.authTokenTTL
}

// GetRefreshTokenTTL returns the refresh token ttl in seconds.
func (c *Config) GetRefreshTokenTTL() int {
	return c.refreshTokenTTL
}

package config_test

import (
	"avito/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testConfig string = `
http-profile:
  host: localhost
  port: 8080
  include-swagger: true
grpc-profile:
  port: 3000
postgres:
  host: 'localhost'
  port: 15432
  connect-attempts: 3
  connect-timeout: 5s
  user: avito
  password: avito
  db: avito
auth:
  jwt:
    tokenTTL: 30m
    issuer: avito.ru
    sign: supersign
`

var testConfigData []byte = []byte(testConfig)

func TestInitConfig_ShouldBindAllConfigParams(t *testing.T) {
	// Arrange
	file := mustWriteConfigToTempFile(t)
	// Ensure the file is cleaned up after the test
	defer os.Remove(file)

	// Act
	cfg, err := config.InitConfig(file)

	// Assert
	require.NoError(t, err, "InitConfig should not return an error")

	require.Equal(t, "localhost", cfg.HTTPConfig.Host, "HTTPConfig.Host should match")
	require.Equal(t, 8080, cfg.HTTPConfig.Port, "HTTPConfig.Port should match")
	require.Equal(t, true, cfg.HTTPConfig.IncludeSwagger, "HTTPConfig.IncludeSwagger should match")

	require.Equal(t, 3000, cfg.GRPCConfig.Port, "GRPCConfig.Port should match")

	require.Equal(t, "localhost", cfg.PostgresConfig.Host, "PostgresConfig.Host should match")
	require.Equal(t, 15432, cfg.PostgresConfig.Port, "PostgresConfig.Port should match")
	require.Equal(t, 3, cfg.PostgresConfig.ConnectAttempts, "PostgresConfig.ConnectAttempts should match")
	require.Equal(t, 5*time.Second, cfg.PostgresConfig.ConnectTimeout, "PostgresConfig.ConnectTimeout should match")
	require.Equal(t, "avito", cfg.PostgresConfig.User, "PostgresConfig.User should match")
	require.Equal(t, "avito", cfg.PostgresConfig.Password, "PostgresConfig.Password should match")
	require.Equal(t, "avito", cfg.PostgresConfig.DB, "PostgresConfig.DB should match")

	require.Equal(t, 30*time.Minute, cfg.AuthConfig.JWTConfig.TokenTTL, "AuthConfig.JWTConfig.TokenTTL should match")
	require.Equal(t, "avito.ru", cfg.AuthConfig.JWTConfig.Issuer, "AuthConfig.JWTConfig.Issuer should match")
	require.Equal(t, "supersign", cfg.AuthConfig.JWTConfig.Sign, "AuthConfig.JWTConfig.Sign should match")
}

func TestInitConfig_ShouldReturnError_WhenNoFilePath(t *testing.T) {
	// Act
	_, err := config.InitConfig("")

	// Assert
	require.Error(t, err)
}

func TestInitConfig_ShouldReturnError_WhenWrongExtenssion(t *testing.T) {
	// Arrange
	const file string = "conf.json"

	// Act
	_, err := config.InitConfig(file)

	// Assert
	require.Error(t, err)
}

func TestInitConfig_ShouldOverrideSomeConfigFromEnv(t *testing.T) {
	//Arrange
	file := mustWriteConfigToTempFile(t)
	const (
		user     string = "MSSQL"
		password string = "MSSQL"
		db       string = "MSSQL"
		sign     string = "MSSQL"
	)
	t.Setenv("POSTGRES_USER", user)
	t.Setenv("POSTGRES_PASSWORD", password)
	t.Setenv("POSTGRES_DB", db)
	t.Setenv("JWT_SIGN", sign)

	// Act
	cfg, err := config.InitConfig(file)

	// Assert
	require.NoError(t, err, "InitConfig should not return an error")

	require.Equal(t, "localhost", cfg.HTTPConfig.Host, "HTTPConfig.Host should match")
	require.Equal(t, 8080, cfg.HTTPConfig.Port, "HTTPConfig.Port should match")
	require.Equal(t, true, cfg.HTTPConfig.IncludeSwagger, "HTTPConfig.IncludeSwagger should match")

	require.Equal(t, 3000, cfg.GRPCConfig.Port, "GRPCConfig.Port should match")

	require.Equal(t, "localhost", cfg.PostgresConfig.Host, "PostgresConfig.Host should match")
	require.Equal(t, 15432, cfg.PostgresConfig.Port, "PostgresConfig.Port should match")
	require.Equal(t, 3, cfg.PostgresConfig.ConnectAttempts, "PostgresConfig.ConnectAttempts should match")
	require.Equal(t, 5*time.Second, cfg.PostgresConfig.ConnectTimeout, "PostgresConfig.ConnectTimeout should match")

	require.Equal(t, 30*time.Minute, cfg.AuthConfig.JWTConfig.TokenTTL, "AuthConfig.JWTConfig.TokenTTL should match")
	require.Equal(t, "avito.ru", cfg.AuthConfig.JWTConfig.Issuer, "AuthConfig.JWTConfig.Issuer should match")

	require.Equal(t, user, cfg.PostgresConfig.User)
	require.Equal(t, password, cfg.PostgresConfig.Password)
	require.Equal(t, db, cfg.PostgresConfig.DB)
	require.Equal(t, sign, cfg.AuthConfig.JWTConfig.Sign)
}

func mustWriteConfigToTempFile(t *testing.T) string {
	t.Helper()

	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%d.yaml", timestamp)
	filePath := filepath.Join(t.TempDir(), fileName)

	if err := os.WriteFile(filePath, testConfigData, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	return filePath
}

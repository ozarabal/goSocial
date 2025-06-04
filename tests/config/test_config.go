package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// TestConfig holds all test configuration
type TestConfig struct {
	API      APIConfig      `json:"api"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Timeouts TimeoutConfig  `json:"timeouts"`
	Parallel ParallelConfig `json:"parallel"`
	Reports  ReportsConfig  `json:"reports"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	BaseURL     string            `json:"base_url"`
	Version     string            `json:"version"`
	Headers     map[string]string `json:"headers"`
	RetryCount  int               `json:"retry_count"`
	RetryDelay  time.Duration     `json:"retry_delay"`
	RateLimits  RateLimitConfig   `json:"rate_limits"`
}

// DatabaseConfig holds database configuration for tests
type DatabaseConfig struct {
	CleanupBetweenTests bool   `json:"cleanup_between_tests"`
	UseTransactions     bool   `json:"use_transactions"`
	SeedData           bool   `json:"seed_data"`
	TestDataPath       string `json:"test_data_path"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	DefaultUsername string        `json:"default_username"`
	DefaultPassword string        `json:"default_password"`
	TokenExpiry     time.Duration `json:"token_expiry"`
	AdminUsername   string        `json:"admin_username"`
	AdminPassword   string        `json:"admin_password"`
}

// TimeoutConfig holds timeout configuration
type TimeoutConfig struct {
	APIRequest      time.Duration `json:"api_request"`
	DatabaseQuery   time.Duration `json:"database_query"`
	TestExecution   time.Duration `json:"test_execution"`
	SetupTeardown   time.Duration `json:"setup_teardown"`
}

// ParallelConfig holds parallel execution configuration
type ParallelConfig struct {
	Enabled        bool `json:"enabled"`
	MaxConcurrency int  `json:"max_concurrency"`
}

// ReportsConfig holds test reporting configuration
type ReportsConfig struct {
	OutputDir      string `json:"output_dir"`
	Format         string `json:"format"` // json, xml, html
	IncludeDetails bool   `json:"include_details"`
	SaveRequests   bool   `json:"save_requests"`
	SaveResponses  bool   `json:"save_responses"`
}

// RateLimitConfig holds rate limiting configuration for testing
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	TestRateLimit     bool          `json:"test_rate_limit"`
	RateLimitDelay    time.Duration `json:"rate_limit_delay"`
}

// Environment represents different test environments
type Environment string

const (
	EnvLocal       Environment = "local"
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// GetTestConfig returns configuration based on environment
func GetTestConfig() *TestConfig {
	env := getEnv("TEST_ENV", "local")
	
	config := &TestConfig{
		API: APIConfig{
			BaseURL:    getEnv("API_BASE_URL", getDefaultBaseURL(Environment(env))),
			Version:    getEnv("API_VERSION", "v1"),
			Headers:    getDefaultHeaders(),
			RetryCount: getEnvInt("API_RETRY_COUNT", 3),
			RetryDelay: getEnvDuration("API_RETRY_DELAY", 1*time.Second),
			RateLimits: RateLimitConfig{
				RequestsPerSecond: getEnvInt("RATE_LIMIT_RPS", 20),
				BurstSize:         getEnvInt("RATE_LIMIT_BURST", 5),
				TestRateLimit:     getEnvBool("TEST_RATE_LIMIT", true),
				RateLimitDelay:    getEnvDuration("RATE_LIMIT_DELAY", 5*time.Second),
			},
		},
		Database: DatabaseConfig{
			CleanupBetweenTests: getEnvBool("DB_CLEANUP_BETWEEN_TESTS", true),
			UseTransactions:     getEnvBool("DB_USE_TRANSACTIONS", true),
			SeedData:           getEnvBool("DB_SEED_DATA", false),
			TestDataPath:       getEnv("TEST_DATA_PATH", "tests/data/fixtures"),
		},
		Auth: AuthConfig{
			DefaultUsername: getEnv("TEST_USERNAME", "testuser@example.com"),
			DefaultPassword: getEnv("TEST_PASSWORD", "password123"),
			TokenExpiry:     getEnvDuration("TOKEN_EXPIRY", 24*time.Hour),
			AdminUsername:   getEnv("ADMIN_USERNAME", "admin@example.com"),
			AdminPassword:   getEnv("ADMIN_PASSWORD", "adminpass123"),
		},
		Timeouts: TimeoutConfig{
			APIRequest:      getEnvDuration("TIMEOUT_API", 30*time.Second),
			DatabaseQuery:   getEnvDuration("TIMEOUT_DB", 10*time.Second),
			TestExecution:   getEnvDuration("TIMEOUT_TEST", 5*time.Minute),
			SetupTeardown:   getEnvDuration("TIMEOUT_SETUP", 30*time.Second),
		},
		Parallel: ParallelConfig{
			Enabled:        getEnvBool("PARALLEL_ENABLED", false),
			MaxConcurrency: getEnvInt("MAX_CONCURRENCY", 4),
		},
		Reports: ReportsConfig{
			OutputDir:      getEnv("REPORTS_DIR", "tests/reports"),
			Format:         getEnv("REPORTS_FORMAT", "json"),
			IncludeDetails: getEnvBool("REPORTS_INCLUDE_DETAILS", true),
			SaveRequests:   getEnvBool("SAVE_REQUESTS", false),
			SaveResponses:  getEnvBool("SAVE_RESPONSES", false),
		},
	}
	
	return config
}

// getDefaultBaseURL returns default base URL for environment
func getDefaultBaseURL(env Environment) string {
	switch env {
	case EnvLocal:
		return "http://localhost:3000/v1"
	case EnvDevelopment:
		return "https://dev-api.gosocial.com/v1"
	case EnvStaging:
		return "https://staging-api.gosocial.com/v1"
	case EnvProduction:
		return "https://api.gosocial.com/v1"
	default:
		return "http://localhost:3000/v1"
	}
}

// getDefaultHeaders returns default headers for all requests
func getDefaultHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   "GoSocial-API-Tests/1.0",
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *TestConfig) Validate() error {
	if c.API.BaseURL == "" {
		return fmt.Errorf("API base URL is required")
	}
	
	if c.API.RetryCount < 0 {
		return fmt.Errorf("API retry count must be non-negative")
	}
	
	if c.Timeouts.APIRequest <= 0 {
		return fmt.Errorf("API timeout must be positive")
	}
	
	if c.Parallel.MaxConcurrency <= 0 {
		c.Parallel.MaxConcurrency = 1
	}
	
	return nil
}

// GetFullURL returns full URL with base URL and endpoint
func (c *TestConfig) GetFullURL(endpoint string) string {
	return c.API.BaseURL + endpoint
}

// IsLocalEnvironment checks if running in local environment
func (c *TestConfig) IsLocalEnvironment() bool {
	return c.API.BaseURL == "http://localhost:3000/v1"
}

// ShouldSkipTest determines if test should be skipped based on environment
func (c *TestConfig) ShouldSkipTest(testType string) bool {
	switch testType {
	case "load":
		return c.IsLocalEnvironment() // Skip load tests in local
	case "security":
		return false // Always run security tests
	case "integration":
		return false // Always run integration tests
	default:
		return false
	}
}
package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/ozarabal/goSocial/tests/config"
	"github.com/ozarabal/goSocial/tests/data/factories"
	"github.com/ozarabal/goSocial/tests/framework/assertions"
	"github.com/ozarabal/goSocial/tests/framework/client"
)

// AuthTestSuite contains all authentication-related tests
type AuthTestSuite struct {
	suite.Suite
	client      *client.APIClient
	config      *config.TestConfig
	userFactory *factories.UserFactory
	valData     *factories.ValidationData
}

// SetupSuite runs once before all tests
func (suite *AuthTestSuite) SetupSuite() {
	suite.config = config.GetTestConfig()
	suite.Require().NoError(suite.config.Validate())
	
	suite.client = client.NewAPIClient(suite.config.API.BaseURL)
	suite.client.SetHeaders(suite.config.API.Headers)
	
	suite.userFactory = factories.NewUserFactory()
	suite.valData = factories.NewValidationData()
}

// SetupTest runs before each test
func (suite *AuthTestSuite) SetupTest() {
	// Add any per-test setup here
}

// TearDownTest runs after each test
func (suite *AuthTestSuite) TearDownTest() {
	// Add any per-test cleanup here
	if suite.config.Database.CleanupBetweenTests {
		// Implement cleanup logic if needed
	}
}

// TestUserRegistration_Success tests successful user registration
func (suite *AuthTestSuite) TestUserRegistration_Success() {
	user := suite.userFactory.Create()
	
	response := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(201).
		ShouldHaveJSONField("data.id").
		ShouldHaveJSONField("data.username").
		ShouldHaveJSONField("data.email").
		ShouldHaveJSONField("data.token").
		ShouldHaveJSONField("data.created_at").
		ShouldHaveJSONField("data.isActive").
		ShouldHaveJSONField("data.role").
		ShouldHaveJSONFieldValue("data.username", user.Username).
		ShouldHaveJSONFieldValue("data.email", user.Email).
		ShouldHaveJSONFieldValue("data.isActive", false). // Should be false until activated
		ShouldHaveJSONFieldType("data.id", "float64").     // JSON numbers are float64
		ShouldHaveJSONFieldType("data.token", "string")
}

// TestUserRegistration_DuplicateEmail tests registration with duplicate email
func (suite *AuthTestSuite) TestUserRegistration_DuplicateEmail() {
	user := suite.userFactory.Create()
	
	// Register user first time
	firstResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), firstResponse).
		ShouldHaveStatus(201)
	
	// Try to register again with same email but different username
	duplicateUser := suite.userFactory.Create()
	duplicateUser.Email = user.Email // Same email
	
	secondResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": duplicateUser.Username,
		"email":    duplicateUser.Email,
		"password": duplicateUser.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), secondResponse).
		ShouldHaveStatus(400).
		ShouldHaveValidErrorResponse()
}

// TestUserRegistration_DuplicateUsername tests registration with duplicate username
func (suite *AuthTestSuite) TestUserRegistration_DuplicateUsername() {
	user := suite.userFactory.Create()
	
	// Register user first time
	firstResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), firstResponse).
		ShouldHaveStatus(201)
	
	// Try to register again with same username but different email
	duplicateUser := suite.userFactory.Create()
	duplicateUser.Username = user.Username // Same username
	
	secondResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": duplicateUser.Username,
		"email":    duplicateUser.Email,
		"password": duplicateUser.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), secondResponse).
		ShouldHaveStatus(400).
		ShouldHaveValidErrorResponse()
}

// TestUserRegistration_ValidationErrors tests various validation errors
func (suite *AuthTestSuite) TestUserRegistration_ValidationErrors() {
	testCases := []struct {
		name     string
		payload  map[string]interface{}
		expected int
	}{
		{
			name: "Missing username",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expected: 400,
		},
		{
			name: "Missing email",
			payload: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			expected: 400,
		},
		{
			name: "Missing password",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
			},
			expected: 400,
		},
		{
			name: "Empty username",
			payload: map[string]interface{}{
				"username": "",
				"email":    "test@example.com",
				"password": "password123",
			},
			expected: 400,
		},
		{
			name: "Invalid email format",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "invalid-email",
				"password": "password123",
			},
			expected: 400,
		},
		{
			name: "Password too short",
			payload: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "12",
			},
			expected: 400,
		},
		{
			name: "Username too long",
			payload: map[string]interface{}{
				"username": string(make([]byte, 101)), // 101 characters
				"email":    "test@example.com",
				"password": "password123",
			},
			expected: 400,
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			response := suite.client.POST("/authentication/user", tc.payload)
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(tc.expected).
				ShouldHaveValidErrorResponse()
		})
	}
}

// TestUserLogin_Success tests successful user login
func (suite *AuthTestSuite) TestUserLogin_Success() {
	// First register a user
	user := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), regResponse).
		ShouldHaveStatus(201)
	
	// Extract activation token and activate user (simplified for test)
	// In real scenario, you'd need to activate via email token
	
	// Now try to login
	loginResponse := suite.client.POST("/authentication/token", map[string]interface{}{
		"email":    user.Email,
		"password": user.Password,
	})
	
	assertions.NewResponseAssertion(suite.T(), loginResponse).
		ShouldHaveStatus(200).
		ShouldHaveValidTokenResponse()
	
	// Verify token format and store it
	token := assertions.NewResponseAssertion(suite.T(), loginResponse).
		GetJSONField("data").(string)
	
	suite.NotEmpty(token, "Token should not be empty")
	suite.client.SetAuth(token)
}

// TestUserLogin_InvalidCredentials tests login with invalid credentials
func (suite *AuthTestSuite) TestUserLogin_InvalidCredentials() {
	testCases := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "Non-existent email",
			email:    "nonexistent@example.com",
			password: "password123",
		},
		{
			name:     "Wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
		},
		{
			name:     "Empty email",
			email:    "",
			password: "password123",
		},
		{
			name:     "Empty password",
			email:    "test@example.com",
			password: "",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			response := suite.client.POST("/authentication/token", map[string]interface{}{
				"email":    tc.email,
				"password": tc.password,
			})
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldBeClientError().
				ShouldHaveValidErrorResponse()
		})
	}
}

// TestUserLogin_ValidationErrors tests login validation errors
func (suite *AuthTestSuite) TestUserLogin_ValidationErrors() {
	invalidEmails := suite.valData.InvalidEmails()
	
	for _, email := range invalidEmails {
		suite.Run("Invalid email: "+email, func() {
			response := suite.client.POST("/authentication/token", map[string]interface{}{
				"email":    email,
				"password": "password123",
			})
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(400).
				ShouldHaveValidErrorResponse()
		})
	}
}

// TestTokenExpiration tests token expiration behavior
func (suite *AuthTestSuite) TestTokenExpiration() {
	// This test would require manipulating time or using a test-specific short-lived token
	// For now, we'll skip this implementation but structure is here
	suite.T().Skip("Token expiration test requires time manipulation - implement with test-specific tokens")
}

// TestSecurityHeaders tests that security headers are present
func (suite *AuthTestSuite) TestSecurityHeaders() {
	user := suite.userFactory.Create()
	
	response := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	})
	
	// Check for security headers
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveHeader("Content-Type")
	
	// Add more security header checks as needed
}

// TestRateLimiting tests API rate limiting
func (suite *AuthTestSuite) TestRateLimiting() {
	if !suite.config.API.RateLimits.TestRateLimit {
		suite.T().Skip("Rate limiting tests disabled")
		return
	}
	
	user := suite.userFactory.Create()
	payload := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"password": user.Password,
	}
	
	// Make requests up to rate limit
	for i := 0; i < suite.config.API.RateLimits.RequestsPerSecond; i++ {
		response := suite.client.POST("/authentication/user", payload)
		// Should succeed or fail for business reasons, not rate limiting
		assertions.NewResponseAssertion(suite.T(), response).
			ShouldSatisfy(func(r *client.APIResponse) bool {
				return r.StatusCode != 429 // Not Too Many Requests
			}, "Should not be rate limited within allowed requests")
	}
	
	// Next request should be rate limited
	response := suite.client.POST("/authentication/user", payload)
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(429)
	
	// Wait for rate limit to reset
	time.Sleep(suite.config.API.RateLimits.RateLimitDelay)
}

// TestSQLInjectionProtection tests protection against SQL injection
func (suite *AuthTestSuite) TestSQLInjectionProtection() {
	maliciousInputs := suite.valData.SpecialCharacterData()
	
	for _, maliciousInput := range maliciousInputs {
		suite.Run("SQL Injection: "+maliciousInput, func() {
			response := suite.client.POST("/authentication/user", map[string]interface{}{
				"username": maliciousInput,
				"email":    maliciousInput + "@example.com",
				"password": "password123",
			})
			
			// Should either succeed with sanitized input or fail with validation error
			// Should NOT cause server error (500)
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldSatisfy(func(r *client.APIResponse) bool {
					return r.StatusCode != 500
				}, "Server should not crash from malicious input")
		})
	}
}

// TestConcurrentRegistrations tests concurrent user registrations
func (suite *AuthTestSuite) TestConcurrentRegistrations() {
	if !suite.config.Parallel.Enabled {
		suite.T().Skip("Parallel tests disabled")
		return
	}
	
	users := suite.userFactory.CreateMultiple(10)
	
	// Channel to collect results
	results := make(chan *client.APIResponse, len(users))
	
	// Start concurrent registrations
	for _, user := range users {
		go func(u *factories.UserData) {
			response := suite.client.POST("/authentication/user", map[string]interface{}{
				"username": u.Username,
				"email":    u.Email,
				"password": u.Password,
			})
			results <- response
		}(user)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < len(users); i++ {
		response := <-results
		if response.StatusCode == 201 {
			successCount++
		}
	}
	
	// All registrations should succeed (assuming unique data)
	suite.Equal(len(users), successCount, "All concurrent registrations should succeed")
}

// Run the test suite
func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
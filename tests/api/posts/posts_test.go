// tests/api/posts/posts_test.go
package posts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ozarabal/goSocial/tests/config"
	"github.com/ozarabal/goSocial/tests/data/factories"
	"github.com/ozarabal/goSocial/tests/framework/assertions"
	"github.com/ozarabal/goSocial/tests/framework/client"
)

// PostsTestSuite contains all posts-related tests
type PostsTestSuite struct {
	suite.Suite
	client       *client.APIClient
	config       *config.TestConfig
	userFactory  *factories.UserFactory
	postFactory  *factories.PostFactory
	valData      *factories.ValidationData
	testUser     *factories.UserData
	authToken    string
}

// SetupSuite runs once before all tests
func (suite *PostsTestSuite) SetupSuite() {
	suite.config = config.GetTestConfig()
	suite.Require().NoError(suite.config.Validate())
	
	suite.client = client.NewAPIClient(suite.config.API.BaseURL)
	suite.client.SetHeaders(suite.config.API.Headers)
	
	suite.userFactory = factories.NewUserFactory()
	suite.postFactory = factories.NewPostFactory()
	suite.valData = factories.NewValidationData()
	
	// Create and authenticate a test user for posts operations
	suite.createAndAuthenticateTestUser()
}

// createAndAuthenticateTestUser creates a user and gets auth token
func (suite *PostsTestSuite) createAndAuthenticateTestUser() {
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
func (suite *PostsTestSuite) SetupTest() {
	// Ensure we have auth token set
	suite.client.SetAuth(suite.authToken)
}

// TestCreatePost_Success tests successful post creation
func (suite *PostsTestSuite) TestCreatePost_Success() {
	post := suite.postFactory.Create()
	
	response := suite.client.POST("/posts", map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"tags":    post.Tags,
	})
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(201).
		ShouldHaveValidPostResponse().
		ShouldHaveJSONFieldValue("data.title", post.Title).
		ShouldHaveJSONFieldValue("data.contetn", post.Content). // Note: typo in original struct
		ShouldHaveJSONFieldValue("data.user_id", float64(suite.testUser.ID)).
		ShouldHaveJSONField("data.id").
		ShouldHaveJSONField("data.created_at").
		ShouldHaveJSONField("data.updated_at").
		ShouldHaveJSONField("data.version").
		ShouldHaveJSONFieldValue("data.version", float64(0)) // New post should have version 0
}

// TestCreatePost_Unauthorized tests post creation without authentication
func (suite *PostsTestSuite) TestCreatePost_Unauthorized() {
	// Remove auth token
	unauthClient := client.NewAPIClient(suite.config.API.BaseURL)
	post := suite.postFactory.Create()
	
	response := unauthClient.POST("/posts", map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"tags":    post.Tags,
	})
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(401).
		ShouldHaveValidErrorResponse()
}

// TestCreatePost_ValidationErrors tests post creation validation
func (suite *PostsTestSuite) TestCreatePost_ValidationErrors() {
	testCases := []struct {
		name     string
		payload  map[string]interface{}
		expected int
	}{
		{
			name: "Missing title",
			payload: map[string]interface{}{
				"content": "Valid content",
				"tags":    []string{"tag1"},
			},
			expected: 400,
		},
		{
			name: "Missing content",
			payload: map[string]interface{}{
				"title": "Valid title",
				"tags":  []string{"tag1"},
			},
			expected: 400,
		},
		{
			name: "Empty title",
			payload: map[string]interface{}{
				"title":   "",
				"content": "Valid content",
				"tags":    []string{"tag1"},
			},
			expected: 400,
		},
		{
			name: "Empty content",
			payload: map[string]interface{}{
				"title":   "Valid title",
				"content": "",
				"tags":    []string{"tag1"},
			},
			expected: 400,
		},
		{
			name: "Title too long",
			payload: map[string]interface{}{
				"title":   string(make([]byte, 101)), // 101 characters (max 100)
				"content": "Valid content",
				"tags":    []string{"tag1"},
			},
			expected: 400,
		},
		{
			name: "Content too long",
			payload: map[string]interface{}{
				"title":   "Valid title",
				"content": string(make([]byte, 1001)), // 1001 characters (max 1000)
				"tags":    []string{"tag1"},
			},
			expected: 400,
		},
		{
			name: "Tags is optional",
			payload: map[string]interface{}{
				"title":   "Valid title",
				"content": "Valid content",
				// No tags - should be OK
			},
			expected: 201,
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			response := suite.client.POST("/posts", tc.payload)
			
			assertion := assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(tc.expected)
			
			if tc.expected == 400 {
				assertion.ShouldHaveValidErrorResponse()
			} else if tc.expected == 201 {
				assertion.ShouldHaveValidPostResponse()
			}
		})
	}
}

// TestGetPost_Success tests successful post retrieval
func (suite *PostsTestSuite) TestGetPost_Success() {
	// First create a post
	post := suite.postFactory.Create()
	createResponse := suite.client.POST("/posts", map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"tags":    post.Tags,
	})
	
	assertions.NewResponseAssertion(suite.T(), createResponse).
		ShouldHaveStatus(201)
	
	// Get the post ID
	postID := assertions.NewResponseAssertion(suite.T(), createResponse).
		GetJSONField("data.id").(float64)
	
	// Now retrieve the post
	getResponse := suite.client.GET(fmt.Sprintf("/posts/%.0f", postID))
	
	assertions.NewResponseAssertion(suite.T(), getResponse).
		ShouldHaveStatus(200).
		ShouldHaveValidPostResponse().
		ShouldHaveJSONFieldValue("data.id", postID).
		ShouldHaveJSONFieldValue("data.title", post.Title).
		ShouldHaveJSONFieldValue("data.contetn", post.Content).
		ShouldHaveJSONField("data.comments") // Should include comments array
}

// TestGetPost_NotFound tests retrieval of non-existent post
func (suite *PostsTestSuite) TestGetPost_NotFound() {
	nonExistentID := 999999
	
	response := suite.client.GET(fmt.Sprintf("/posts/%d", nonExistentID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(404).
		ShouldHaveValidErrorResponse()
}

// TestGetPost_InvalidID tests retrieval with invalid post ID
func (suite *PostsTestSuite) TestGetPost_InvalidID() {
	invalidIDs := []string{"abc", "0", "-1", "1.5"}
	
	for _, id := range invalidIDs {
		suite.Run("Invalid ID: "+id, func() {
			response := suite.client.GET("/posts/" + id)
			
			// Should be either 400 (bad request) or 500 (server error)
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldSatisfy(func(r *client.APIResponse) bool {
					return r.StatusCode == 400 || r.StatusCode == 500
				}, "Should return error for invalid ID")
		})
	}
}

// TestUpdatePost_Success tests successful post update
func (suite *PostsTestSuite) TestUpdatePost_Success() {
	// Create a post first
	originalPost := suite.postFactory.Create()
	createResponse := suite.client.POST("/posts", map[string]interface{}{
		"title":   originalPost.Title,
		"content": originalPost.Content,
		"tags":    originalPost.Tags,
	})
	
	assertions.NewResponseAssertion(suite.T(), createResponse).
		ShouldHaveStatus(201)
	
	postID := assertions.NewResponseAssertion(suite.T(), createResponse).
		GetJSONField("data.id").(float64)
	
	// Update the post
	updatedPost := suite.postFactory.Create()
	updateResponse := suite.client.PATCH(fmt.Sprintf("/posts/%.0f", postID), map[string]interface{}{
		"title":   updatedPost.Title,
		"content": updatedPost.Content,
	})
	
	assertions.NewResponseAssertion(suite.T(), updateResponse).
		ShouldHaveStatus(200).
		ShouldHaveValidPostResponse().
		ShouldHaveJSONFieldValue("data.title", updatedPost.Title).
		ShouldHaveJSONFieldValue("data.contetn", updatedPost.Content).
		ShouldHaveJSONFieldValue("data.version", float64(1)) // Version should increment
}

// TestUpdatePost_PartialUpdate tests partial post updates
func (suite *PostsTestSuite) TestUpdatePost_PartialUpdate() {
	// Create a post
	originalPost := suite.postFactory.Create()
	createResponse := suite.client.POST("/posts", map[string]interface{}{
		"title":   originalPost.Title,
		"content": originalPost.Content,
		"tags":    originalPost.Tags,
	})
	
	postID := assertions.NewResponseAssertion(suite.T(), createResponse).
		GetJSONField("data.id").(float64)
	
	// Update only title
	newTitle := "Updated Title Only"
	updateResponse := suite.client.PATCH(fmt.Sprintf("/posts/%.0f", postID), map[string]interface{}{
		"title": newTitle,
	})
	
	assertions.NewResponseAssertion(suite.T(), updateResponse).
		ShouldHaveStatus(200).
		ShouldHaveJSONFieldValue("data.title", newTitle).
		ShouldHaveJSONFieldValue("data.contetn", originalPost.Content) // Content should remain unchanged
}

// TestUpdatePost_Unauthorized tests updating post without proper permissions
func (suite *PostsTestSuite) TestUpdatePost_Unauthorized() {
	// Create a post with current user
	post := suite.postFactory.Create()
	createResponse := suite.client.POST("/posts", map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"tags":    post.Tags,
	})
	
	postID := assertions.NewResponseAssertion(suite.T(), createResponse).
		GetJSONField("data.id").(float64)
	
	// Create another user and try to update the post
	anotherUser := suite.userFactory.Create()
	suite.client.POST("/authentication/user", map[string]interface{}{
		"username": anotherUser.Username,
		"email":    anotherUser.Email,
		"password": anotherUser.Password,
	})
	
	loginResponse := suite.client.POST("/authentication/token", map[string]interface{}{
		"email":    anotherUser.Email,
		"password": anotherUser.Password,
	})
	
	anotherToken := assertions.NewResponseAssertion(suite.T(), loginResponse).
		GetJSONField("data").(string)
	
	// Try to update with different user's token
	anotherClient := client.NewAPIClient(suite.config.API.BaseURL)
	anotherClient.SetAuth(anotherToken)
	
	updateResponse := anotherClient.PATCH(fmt.Sprintf("/posts/%.0f", postID), map[string]interface{}{
		"title": "Unauthorized update attempt",
	})
	
	// Should be forbidden (403) or unauthorized (401)
	assertions.NewResponseAssertion(suite.T(), updateResponse).
		ShouldSatisfy(func(r *client.APIResponse) bool {
			return r.StatusCode == 401 || r.StatusCode == 403
		}, "Should not allow unauthorized post updates")
}

// TestDeletePost_Success tests successful post deletion
func (suite *PostsTestSuite) TestDeletePost_Success() {
	// Create a post
	post := suite.postFactory.Create()
	createResponse := suite.client.POST("/posts", map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"tags":    post.Tags,
	})
	
	postID := assertions.NewResponseAssertion(suite.T(), createResponse).
		GetJSONField("data.id").(float64)
	
	// Delete the post
	deleteResponse := suite.client.DELETE(fmt.Sprintf("/posts/%.0f", postID))
	
	assertions.NewResponseAssertion(suite.T(), deleteResponse).
		ShouldHaveStatus(204).
		ShouldHaveEmptyBody()
	
	// Verify post is deleted by trying to get it
	getResponse := suite.client.GET(fmt.Sprintf("/posts/%.0f", postID))
	assertions.NewResponseAssertion(suite.T(), getResponse).
		ShouldHaveStatus(404)
}

// TestDeletePost_NotFound tests deletion of non-existent post
func (suite *PostsTestSuite) TestDeletePost_NotFound() {
	nonExistentID := 999999
	
	response := suite.client.DELETE(fmt.Sprintf("/posts/%d", nonExistentID))
	
	assertions.NewResponseAssertion(suite.T(), response).
		ShouldHaveStatus(404).
		ShouldHaveValidErrorResponse()
}

// TestPostWithComments tests post retrieval with comments
func (suite *PostsTestSuite) TestPostWithComments() {
	// Create a post
	post := suite.postFactory.Create()
	createResponse := suite.client.POST("/posts", map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"tags":    post.Tags,
	})
	
	postID := assertions.NewResponseAssertion(suite.T(), createResponse).
		GetJSONField("data.id").(float64)
	
	// Get the post (should include empty comments array)
	getResponse := suite.client.GET(fmt.Sprintf("/posts/%.0f", postID))
	
	assertions.NewResponseAssertion(suite.T(), getResponse).
		ShouldHaveStatus(200).
		ShouldHaveJSONField("data.comments").
		ShouldHaveJSONArrayLength("data.comments", 0) // Should be empty initially
}

// TestCreateMultiplePosts tests creating multiple posts
func (suite *PostsTestSuite) TestCreateMultiplePosts() {
	posts := suite.postFactory.CreateMultiple(5)
	createdPosts := make([]float64, 0, len(posts))
	
	for i, post := range posts {
		suite.Run(fmt.Sprintf("Create post %d", i+1), func() {
			response := suite.client.POST("/posts", map[string]interface{}{
				"title":   post.Title,
				"content": post.Content,
				"tags":    post.Tags,
			})
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(201).
				ShouldHaveValidPostResponse()
			
			postID := assertions.NewResponseAssertion(suite.T(), response).
				GetJSONField("data.id").(float64)
			
			createdPosts = append(createdPosts, postID)
		})
	}
	
	// Verify all posts were created with unique IDs
	suite.Equal(len(posts), len(createdPosts), "All posts should be created")
	
	// Check for unique IDs
	idMap := make(map[float64]bool)
	for _, id := range createdPosts {
		suite.False(idMap[id], "Post IDs should be unique")
		idMap[id] = true
	}
}

// TestPostSecurityInjection tests security against injection attacks
func (suite *PostsTestSuite) TestPostSecurityInjection() {
	maliciousInputs := suite.valData.SpecialCharacterData()
	
	for _, maliciousInput := range maliciousInputs {
		suite.Run("Security test: "+maliciousInput, func() {
			response := suite.client.POST("/posts", map[string]interface{}{
				"title":   maliciousInput,
				"content": maliciousInput,
				"tags":    []string{maliciousInput},
			})
			
			// Should either succeed (with sanitized input) or fail with validation error
			// Should NOT cause server error
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldSatisfy(func(r *client.APIResponse) bool {
					return r.StatusCode != 500
				}, "Server should handle malicious input gracefully")
		})
	}
}

// TestConcurrentPostOperations tests concurrent post operations
func (suite *PostsTestSuite) TestConcurrentPostOperations() {
	if !suite.config.Parallel.Enabled {
		suite.T().Skip("Parallel tests disabled")
		return
	}
	
	posts := suite.postFactory.CreateMultiple(10)
	results := make(chan *client.APIResponse, len(posts))
	
	// Create posts concurrently
	for _, post := range posts {
		go func(p *factories.PostData) {
			response := suite.client.POST("/posts", map[string]interface{}{
				"title":   p.Title,
				"content": p.Content,
				"tags":    p.Tags,
			})
			results <- response
		}(post)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < len(posts); i++ {
		response := <-results
		if response.StatusCode == 201 {
			successCount++
		}
	}
	
	suite.Equal(len(posts), successCount, "All concurrent post creations should succeed")
}

// TestPostTagsHandling tests various tag scenarios
func (suite *PostsTestSuite) TestPostTagsHandling() {
	testCases := []struct {
		name string
		tags []string
	}{
		{
			name: "No tags",
			tags: []string{},
		},
		{
			name: "Single tag",
			tags: []string{"technology"},
		},
		{
			name: "Multiple tags",
			tags: []string{"tech", "golang", "api"},
		},
		{
			name: "Tags with spaces",
			tags: []string{"tech news", "social media"},
		},
		{
			name: "Duplicate tags",
			tags: []string{"tech", "tech", "golang"},
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			post := suite.postFactory.Create()
			response := suite.client.POST("/posts", map[string]interface{}{
				"title":   post.Title,
				"content": post.Content,
				"tags":    tc.tags,
			})
			
			assertions.NewResponseAssertion(suite.T(), response).
				ShouldHaveStatus(201).
				ShouldHaveJSONField("data.tags")
		})
	}
}

// Run the test suite
func TestPostsTestSuite(t *testing.T) {
	suite.Run(t, new(PostsTestSuite))
}
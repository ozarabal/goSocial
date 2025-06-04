package assertions

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ozarabal/goSocial/tests/framework/client"
)

type ResponseAssertion struct {
	t        *testing.T
	response *client.APIResponse
	failed   bool
}

func NewResponseAssertion(t *testing.T, response *client.APIResponse) *ResponseAssertion {
	return &ResponseAssertion{
		t:        t,
		response: response,
		failed:   false,
	}
}

func (ra *ResponseAssertion) ShouldHaveStatus(expectedStatus int) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	if !assert.Equal(ra.t, expectedStatus, ra.response.StatusCode, 
		"Expected status %d, got %d. Response: %s", 
		expectedStatus, ra.response.StatusCode, string(ra.response.Body)) {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldBeSuccessful() *ResponseAssertion {
	return ra.ShouldHaveStatusInRange(200, 299)
}

func (ra *ResponseAssertion) ShouldBeClientError() *ResponseAssertion {
	return ra.ShouldHaveStatusInRange(400, 499)
}

func (ra *ResponseAssertion) ShouldBeServerError() *ResponseAssertion {
	return ra.ShouldHaveStatusInRange(500, 599)
}

func (ra *ResponseAssertion) ShouldHaveStatusInRange(min, max int) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	if ra.response.StatusCode < min || ra.response.StatusCode > max {
		ra.t.Errorf("Expected status in range %d-%d, got %d", min, max, ra.response.StatusCode)
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveHeader(headerName string) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	if ra.response.Headers == nil {
		ra.t.Errorf("Response headers are nil")
		ra.failed = true
		return ra
	}
	
	if !assert.Contains(ra.t, ra.response.Headers, headerName, 
		"Response should contain header: %s", headerName) {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveHeaderValue(headerName, expectedValue string) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	if ra.response.Headers == nil {
		ra.t.Errorf("Response headers are nil")
		ra.failed = true
		return ra
	}
	
	headerValue := ra.response.Headers.Get(headerName)
	if !assert.Equal(ra.t, expectedValue, headerValue, 
		"Header %s should have value %s, got %s", headerName, expectedValue, headerValue) {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveBodyContaining(substring string) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	bodyStr := string(ra.response.Body)
	if !assert.Contains(ra.t, bodyStr, substring, 
		"Response body should contain: %s", substring) {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveEmptyBody() *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	if !assert.Empty(ra.t, ra.response.Body, "Response body should be empty") {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveJSONField(fieldPath string) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	value, exists := ra.getJSONFieldValue(fieldPath)
	if !assert.True(ra.t, exists, "JSON field '%s' should exist. Response: %s", 
		fieldPath, string(ra.response.Body)) {
		ra.failed = true
	}
	_ = value
	return ra
}

func (ra *ResponseAssertion) ShouldHaveJSONFieldValue(fieldPath string, expectedValue interface{}) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	value, exists := ra.getJSONFieldValue(fieldPath)
	if !exists {
		ra.t.Errorf("JSON field '%s' does not exist", fieldPath)
		ra.failed = true
		return ra
	}
	
	if !assert.Equal(ra.t, expectedValue, value, 
		"JSON field '%s' should have value %v, got %v", fieldPath, expectedValue, value) {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveJSONFieldType(fieldPath string, expectedType string) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	value, exists := ra.getJSONFieldValue(fieldPath)
	if !exists {
		ra.t.Errorf("JSON field '%s' does not exist", fieldPath)
		ra.failed = true
		return ra
	}
	
	if value == nil {
		ra.t.Errorf("JSON field '%s' is nil", fieldPath)
		ra.failed = true
		return ra
	}
	
	actualType := reflect.TypeOf(value).String()
	if !strings.Contains(actualType, expectedType) {
		ra.t.Errorf("JSON field '%s' should be type %s, got %s", fieldPath, expectedType, actualType)
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveJSONArrayLength(fieldPath string, expectedLength int) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	value, exists := ra.getJSONFieldValue(fieldPath)
	if !exists {
		ra.t.Errorf("JSON field '%s' does not exist", fieldPath)
		ra.failed = true
		return ra
	}
	
	arrayValue, ok := value.([]interface{})
	if !ok {
		ra.t.Errorf("JSON field '%s' is not an array", fieldPath)
		ra.failed = true
		return ra
	}
	
	if !assert.Equal(ra.t, expectedLength, len(arrayValue), 
		"JSON array '%s' should have length %d, got %d", fieldPath, expectedLength, len(arrayValue)) {
		ra.failed = true
	}
	return ra
}

func (ra *ResponseAssertion) ShouldHaveValidUserResponse() *ResponseAssertion {
	return ra.ShouldHaveJSONField("data.id").
		ShouldHaveJSONField("data.username").
		ShouldHaveJSONField("data.email").
		ShouldHaveJSONField("data.created_at").
		ShouldHaveJSONField("data.isActive").
		ShouldHaveJSONField("data.role")
}

func (ra *ResponseAssertion) ShouldHaveValidPostResponse() *ResponseAssertion {
	return ra.ShouldHaveJSONField("data.id").
		ShouldHaveJSONField("data.title").
		ShouldHaveJSONField("data.contetn").
		ShouldHaveJSONField("data.user_id").
		ShouldHaveJSONField("data.created_at").
		ShouldHaveJSONField("data.tags")
}

func (ra *ResponseAssertion) ShouldHaveValidTokenResponse() *ResponseAssertion {
	return ra.ShouldHaveJSONField("data").
		ShouldSatisfy(func(r *client.APIResponse) bool {
			tokenData := ra.GetJSONField("data")
			return tokenData != nil && tokenData != ""
		}, "Token response should have non-empty data field")
}

func (ra *ResponseAssertion) ShouldHaveValidErrorResponse() *ResponseAssertion {
	return ra.ShouldHaveJSONField("error").
		ShouldHaveJSONFieldType("error", "string")
}

func (ra *ResponseAssertion) getJSONFieldValue(fieldPath string) (interface{}, bool) {
	if ra.response.JSON == nil {
		return nil, false
	}
	
	fields := strings.Split(fieldPath, ".")
	var current interface{} = ra.response.JSON
	
	for _, field := range fields {
		if currentMap, ok := current.(map[string]interface{}); ok {
			if value, exists := currentMap[field]; exists {
				current = value
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	
	return current, true
}

func (ra *ResponseAssertion) GetJSONField(fieldPath string) interface{} {
	value, _ := ra.getJSONFieldValue(fieldPath)
	return value
}

func (ra *ResponseAssertion) Debug() *ResponseAssertion {
	fmt.Printf("=== RESPONSE DEBUG ===\n")
	fmt.Printf("Status: %d\n", ra.response.StatusCode)
	if ra.response.Headers != nil {
		fmt.Printf("Headers: %+v\n", ra.response.Headers)
	} else {
		fmt.Printf("Headers: nil\n")
	}
	fmt.Printf("Body: %s\n", string(ra.response.Body))
	fmt.Printf("JSON: %+v\n", ra.response.JSON)
	fmt.Printf("Error: %v\n", ra.response.Error)
	fmt.Printf("=====================\n")
	return ra
}

func (ra *ResponseAssertion) ShouldSatisfy(validationFunc func(*client.APIResponse) bool, message string) *ResponseAssertion {
	if ra.failed {
		return ra
	}
	
	if !validationFunc(ra.response) {
		ra.t.Error(message)
		ra.failed = true
	}
	return ra
}
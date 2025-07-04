package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/bernaotieno/social-network/backend/pkg/db/sqlite"
	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/google/uuid"
)

func TestPrivatePostVisibility(t *testing.T) {
	// Initialize test database
	db, err := sqlite.NewDB("./test_social_network.db")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := sqlite.RunMigrations("./test_social_network.db", "./pkg/db/migrations/sqlite"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test services
	postService := models.NewPostService(db)

	// Create test users
	user1 := &models.User{
		ID:       uuid.New().String(),
		Username: "testuser1",
		Email:    "test1@example.com",
		FullName: "Test User 1",
		Password: "hashedpassword1",
	}

	user2 := &models.User{
		ID:       uuid.New().String(),
		Username: "testuser2",
		Email:    "test2@example.com",
		FullName: "Test User 2",
		Password: "hashedpassword2",
	}

	// Insert users directly into database
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, full_name, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user1.ID, user1.Username, user1.Email, user1.FullName, user1.Password, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, email, full_name, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user2.ID, user2.Username, user2.Email, user2.FullName, user2.Password, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	fmt.Printf("Created test users:\n")
	fmt.Printf("User 1: %s (%s)\n", user1.Username, user1.ID)
	fmt.Printf("User 2: %s (%s)\n", user2.Username, user2.ID)

	// Create test posts with different visibility levels
	publicPost := &models.Post{
		UserID:     user1.ID,
		Content:    "This is a public post from user1",
		Visibility: models.PostVisibilityPublic,
	}

	privatePost := &models.Post{
		UserID:     user1.ID,
		Content:    "This is a PRIVATE post from user1 - should only be visible to user1",
		Visibility: models.PostVisibilityPrivate,
	}

	followersPost := &models.Post{
		UserID:     user1.ID,
		Content:    "This is a followers-only post from user1",
		Visibility: models.PostVisibilityFollowers,
	}

	// Create posts
	if err := postService.Create(publicPost); err != nil {
		t.Fatalf("Failed to create public post: %v", err)
	}

	if err := postService.Create(privatePost); err != nil {
		t.Fatalf("Failed to create private post: %v", err)
	}

	if err := postService.Create(followersPost); err != nil {
		t.Fatalf("Failed to create followers post: %v", err)
	}

	fmt.Printf("\nCreated test posts:\n")
	fmt.Printf("Public post: %s\n", publicPost.ID)
	fmt.Printf("Private post: %s\n", privatePost.ID)
	fmt.Printf("Followers post: %s\n", followersPost.ID)

	// Test 1: User1 viewing their own posts (should see all 3)
	fmt.Printf("\n=== TEST 1: User1 viewing their own posts ===\n")
	user1Posts, err := postService.GetUserPosts(user1.ID, user1.ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get user1's posts: %v", err)
	}
	fmt.Printf("User1 can see %d of their own posts:\n", len(user1Posts))
	for _, post := range user1Posts {
		fmt.Printf("  - %s (%s)\n", post.Content, post.Visibility)
	}

	if len(user1Posts) != 3 {
		t.Errorf("Expected user1 to see 3 posts, but got %d", len(user1Posts))
	}

	// Test 2: User2 viewing User1's posts (should only see public post)
	fmt.Printf("\n=== TEST 2: User2 viewing User1's posts (not following) ===\n")
	user1PostsFromUser2, err := postService.GetUserPosts(user1.ID, user2.ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get user1's posts from user2 perspective: %v", err)
	}
	fmt.Printf("User2 can see %d of User1's posts:\n", len(user1PostsFromUser2))
	for _, post := range user1PostsFromUser2 {
		fmt.Printf("  - %s (%s)\n", post.Content, post.Visibility)
	}

	// Verify the fix: User2 should NOT see the private post
	privatePostVisible := false
	for _, post := range user1PostsFromUser2 {
		if post.Visibility == models.PostVisibilityPrivate {
			privatePostVisible = true
			break
		}
	}

	fmt.Printf("\n=== VERIFICATION RESULTS ===\n")
	if privatePostVisible {
		fmt.Printf("❌ FAILED: Private post is visible to other users!\n")
		t.Error("Private post should not be visible to other users")
	} else {
		fmt.Printf("✅ SUCCESS: Private post is properly hidden from other users!\n")
	}

	// User2 should only see the public post (1 post)
	if len(user1PostsFromUser2) != 1 {
		t.Errorf("Expected user2 to see 1 post (public only), but got %d", len(user1PostsFromUser2))
	}

	// Test 3: Check individual post access
	fmt.Printf("\n=== TEST 3: Individual post access ===\n")

	// User2 trying to access private post directly
	_, err = postService.GetByID(privatePost.ID, user2.ID)
	if err != nil {
		fmt.Printf("✅ SUCCESS: User2 cannot access private post directly (Error: %s)\n", err.Error())
	} else {
		fmt.Printf("❌ FAILED: User2 can access private post directly!\n")
		t.Error("User2 should not be able to access private post directly")
	}

	// User1 accessing their own private post
	_, err = postService.GetByID(privatePost.ID, user1.ID)
	if err != nil {
		fmt.Printf("❌ FAILED: User1 cannot access their own private post (Error: %s)\n", err.Error())
		t.Error("User1 should be able to access their own private post")
	} else {
		fmt.Printf("✅ SUCCESS: User1 can access their own private post\n")
	}

	fmt.Printf("\n=== TEST COMPLETED ===\n")
}

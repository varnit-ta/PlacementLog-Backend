package posts

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/varnit-ta/PlacementLog/internal/db"
)

/*
PostsRepo handles post-related data access operations.
Provides methods for creating, reading, updating, and deleting posts in the database.
*/
type PostsRepo struct {
	db *sql.DB
}

/*
NewPostsRepo creates a new PostsRepo instance with the provided database connection.

Parameters:
- db: The database connection

Returns:
- *PostsRepo: A new repository instance
*/
func NewPostsRepo(db *sql.DB) *PostsRepo {
	return &PostsRepo{
		db: db,
	}
}

/*
GetAllPosts retrieves all approved posts from the database.
Only returns posts that have been reviewed and approved by admins.

Returns:
- []db.Post: List of all approved posts
- error: Any error that occurred during retrieval

The function queries the database for posts where reviewed=true.
*/
func (repo PostsRepo) GetAllPosts() ([]db.Post, error) {
	query := `
		SELECT id, user_id, post_body, reviewed
		FROM placement_log_posts
		WHERE reviewed=true;
	`

	rows, err := repo.db.Query(query)

	if err != nil {
		return nil, fmt.Errorf("failed to get all the posts: %v", err)
	}

	defer rows.Close()

	var posts []db.Post

	for rows.Next() {
		var p db.Post
		var reviewed bool
		if err := rows.Scan(&p.ID, &p.UserID, &p.PostBody, &reviewed); err != nil {
			return nil, fmt.Errorf("failed to scan posts: %v", err)
		}
		p.Reviewed = reviewed
		posts = append(posts, p)
	}

	return posts, nil
}

/*
GetAllPostsForAdmin retrieves all posts from the database for admin review.
Returns both approved and pending posts, ordered by creation date.

Returns:
- []db.Post: List of all posts (approved and pending)
- error: Any error that occurred during retrieval

The function queries the database for all posts, including the reviewed status.
Posts are ordered by created_at in descending order (newest first).
*/
func (repo PostsRepo) GetAllPostsForAdmin() ([]db.Post, error) {
	query := `
		SELECT id, user_id, post_body, reviewed
		FROM placement_log_posts
		ORDER BY created_at DESC;
	`

	rows, err := repo.db.Query(query)

	if err != nil {
		return nil, fmt.Errorf("failed to get all posts for admin: %v", err)
	}

	defer rows.Close()

	var posts []db.Post

	for rows.Next() {
		var p db.Post
		var reviewed bool
		if err := rows.Scan(&p.ID, &p.UserID, &p.PostBody, &reviewed); err != nil {
			return nil, fmt.Errorf("failed to scan posts: %v", err)
		}
		p.Reviewed = reviewed
		posts = append(posts, p)
	}

	return posts, nil
}

/*
GetPostsByUserId retrieves approved posts for a specific user.

Parameters:
- userId: The ID of the user whose posts to retrieve

Returns:
- []db.Post: List of the user's approved posts
- error: Any error that occurred during retrieval

The function queries the database for posts that belong to the specified user
and have been reviewed and approved (reviewed=true).

Possible errors:
- "all fields are required": Missing user ID
- "failed to get user posts": Database query error
*/
func (repo PostsRepo) GetPostsByUserId(userId string) ([]db.Post, error) {
	if userId == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	query := `
		SELECT id, user_id, post_body, reviewed
		FROM placement_log_posts 
		WHERE user_id=$1 AND reviewed=true;`

	rows, err := repo.db.Query(query, userId)

	if err != nil {
		return nil, fmt.Errorf("failed to get user posts: %v", err)
	}

	defer rows.Close()

	var posts []db.Post

	for rows.Next() {
		var p db.Post
		var reviewed bool

		if err := rows.Scan(&p.ID, &p.UserID, &p.PostBody, &reviewed); err != nil {
			return nil, fmt.Errorf("failed to scan posts: %v", err)
		}

		p.Reviewed = reviewed
		posts = append(posts, p)
	}

	return posts, nil
}

/*
AddPost creates a new post in the database.
The post is created with reviewed=false, requiring admin approval.

Parameters:
- userId: The ID of the user creating the post
- postBody: The post content as JSON

Returns:
- *db.Post: The created post
- error: Any error that occurred during creation

The function:
1. Validates that user ID and post body are provided
2. Inserts the new post into the database with reviewed=false
3. Returns the created post information

Possible errors:
- "all fields are required": Missing user ID or post body
- "failed to add post": Database insertion error
*/
func (repo PostsRepo) AddPost(userId string, postBody json.RawMessage) (*db.Post, error) {
	if userId == "" || postBody == nil {
		return nil, fmt.Errorf("all fields are required")
	}

	query := `
		INSERT INTO placement_log_posts (user_id, post_body, reviewed)
		VALUES ($1, $2, false)
		RETURNING id, user_id, post_body, reviewed;
	`

	var post db.Post
	err := repo.db.QueryRow(query, userId, postBody).Scan(&post.ID, &post.UserID, &post.PostBody, &post.Reviewed)

	if err != nil {
		return nil, fmt.Errorf("failed to add post: %v", err)
	}

	return &post, nil
}

/*
UpdatePost updates an existing post in the database.
Users can only update their own posts. When updated, the post needs re-review.

Parameters:
- postId: The ID of the post to update
- userId: The ID of the user updating the post
- postBody: The updated post content as JSON

Returns:
- *db.Post: The updated post
- error: Any error that occurred during update

The function:
1. Validates that post ID, user ID, and post body are provided
2. Updates the post in the database (sets reviewed=false)
3. Returns the updated post information

Possible errors:
- "all fields are required": Missing required parameters
- "post not found or unauthorized": Post doesn't exist or user doesn't own it
- "failed to update post": Database update error

Note: When a post is updated, it needs to be reviewed again by an admin.
*/
func (repo PostsRepo) UpdatePost(postId string, userId string, postBody json.RawMessage) (*db.Post, error) {
	if postId == "" || userId == "" || postBody == nil {
		return nil, fmt.Errorf("all fields are required")
	}

	query := `
		UPDATE placement_log_posts 
		SET post_body = $1, reviewed = false
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, post_body, reviewed;
	`

	var post db.Post
	err := repo.db.QueryRow(query, postBody, postId, userId).Scan(&post.ID, &post.UserID, &post.PostBody, &post.Reviewed)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found or unauthorized")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update post: %v", err)
	}

	return &post, nil
}

/*
DeletePost deletes a post from the database.
Users can only delete their own posts.

Parameters:
- postId: The ID of the post to delete
- userId: The ID of the user deleting the post

Returns:
- error: Any error that occurred during deletion

The function:
1. Validates that post ID and user ID are provided
2. Deletes the post from the database (only if owned by the user)
3. Returns any error that occurred

Possible errors:
- "post ID and user ID are required": Missing required parameters
- "failed to delete post": Database deletion error
- "no post found with given ID or unauthorized": Post doesn't exist or user doesn't own it
*/
func (repo PostsRepo) DeletePost(postId string, userId string) error {
	if postId == "" || userId == "" {
		return fmt.Errorf("post ID and user ID are required")
	}

	query := `
		DELETE FROM placement_log_posts
		WHERE id = $1 AND user_id = $2;
	`

	result, err := repo.db.Exec(query, postId, userId)

	if err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not confirm deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no post found with given ID or unauthorized")
	}

	return nil
}

/*
DeletePostAsAdmin deletes any post from the database (admin operation).
Admins can delete any post, regardless of ownership.

Parameters:
- postId: The ID of the post to delete

Returns:
- error: Any error that occurred during deletion

The function:
1. Validates that post ID is provided
2. Deletes the post from the database
3. Returns any error that occurred

Possible errors:
- "post ID is required": Missing post ID
- "failed to delete post": Database deletion error
- "no post found with given ID": Post doesn't exist
*/
func (repo PostsRepo) DeletePostAsAdmin(postId string) error {
	if postId == "" {
		return fmt.Errorf("post ID is required")
	}

	query := `
		DELETE FROM placement_log_posts
		WHERE id = $1;
	`

	result, err := repo.db.Exec(query, postId)

	if err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not confirm deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no post found with given ID")
	}

	return nil
}

/*
ReviewPost reviews a post (admin operation).
Admins can approve or reject posts by updating the reviewed status.

Parameters:
- postId: The ID of the post to review
- action: The review action ("approve" or "reject")

Returns:
- error: Any error that occurred during review

The function:
1. Validates that post ID and action are provided
2. Updates the post's reviewed status in the database
3. Returns any error that occurred

Possible errors:
- "post ID and action are required": Missing required parameters
- "invalid action: must be 'approve' or 'reject'": Invalid action value
- "failed to review post": Database update error
- "no post found with given ID": Post doesn't exist

Note: When a post is approved (reviewed=true), it becomes visible to the public.
When a post is rejected (reviewed=false), it remains hidden from public view.
*/
func (repo PostsRepo) ReviewPost(postId string, action string) error {
	if postId == "" || action == "" {
		return fmt.Errorf("post ID and action are required")
	}

	var reviewed bool
	switch action {
	case "approve":
		reviewed = true
	case "reject":
		reviewed = false
	default:
		return fmt.Errorf("invalid action: must be 'approve' or 'reject'")
	}

	query := `
		UPDATE placement_log_posts 
		SET reviewed = $1
		WHERE id = $2;
	`

	result, err := repo.db.Exec(query, reviewed, postId)

	if err != nil {
		return fmt.Errorf("failed to review post: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not confirm review: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no post found with given ID")
	}

	return nil
}

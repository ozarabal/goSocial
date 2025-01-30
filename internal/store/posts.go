package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)
type Post struct{
	ID			int64	`json:"id"`
	Content 	string	`json:"contetn"`
	Title		string	`json:"title"`
	UserId		int64	`json:"user_id"`
	Tags		[]string`json:"tags"`
	CreatedAt	string	`json:"created_at"`
	UpdatedAt	string	`json:"updated_at"`
	Version 	int		`json:"version"`
	Comments	[]Comment`json:"comments"`
	User		User	`json:"user"`
}

type PostWithMetadata struct{
	Post 
	CommentCount int `json:"comments_count"`
}

type PostsStorage struct {
	db *sql.DB
}

func (s *PostsStorage) Create(ctx context.Context, post *Post) error{
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil{
		return err
	}

	return nil
}

func (s *PostsStorage) GetByID(ctx context.Context, id int64) (*Post,error) {
	query := `
		SELECT id, user_id, title, content, created_at::TEXT AS created_at, updated_at, tags, version
		FROM posts
		WHERE id = $1 
	`
	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,	
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err != nil {
		switch{
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		
		default:
			return nil, err
		}
	}

	
	return &post, nil
}

func (s *PostsStorage) DeleteByID(ctx context.Context, id int64) error{
	query := `
		DELETE FROM post
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrorNotFound
	}

	return nil
}

func (s *PostsStorage) Update(ctx context.Context, post *Post) error{
	query := `
		UPDATE posts
		SET title = $1, content = $2, version = version+1
		WHERE id = $3 and version = $4
		RETURNING version
	`

	err :=  s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil{
		switch  {
		case errors.Is(err, sql.ErrNoRows):
			return ErrorNotFound
		default:
			return err
		}
	}
	return nil
}

func (s* PostsStorage) GetUserFeed(ctx context.Context, id int64, param PaginatedFeedQuery) ([]PostWithMetadata,error){
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags, COUNT(c.id) AS comments_count
		, u.username
		FROM posts p 
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN followers f ON f.user_id = p.user_id
		WHERE (f.follower_id = $1 OR p.user_id = $1) AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
		(p.tags @> $5 OR $5 = '{}')
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` +param.Sort+ `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, id, param.Limit, param.Offset, param.Search, pq.Array(param.Tags))
	if err != nil {
		return nil ,err
	}
	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserId,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.CommentCount,
			&p.User.Username,
		)

		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, err
	
	
}
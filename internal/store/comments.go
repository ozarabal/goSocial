package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostId    int64  `json:"post_id"`
	UserId    int64  `json:"user_id"`
	Content   string  `json:"content"`
	CreatedAt string `json:"created_at"`
	User		User	`json:"user"`
}

type CommentsStorage struct {
	db *sql.DB
}

func (s *CommentsStorage) Create(ctx context.Context, comment *Comment) error{
	query := `
		INSERT INTO comments(content, post_id, user_id)
		VALUES ($1,$2,$3) RETURNING id, created_at 
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.Content,
		comment.PostId,
		comment.UserId,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *CommentsStorage) GetCommentsByPostID(ctx context.Context, id int64) ( []Comment,error){
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.id
		FROM comments c join users u on u.id = c.user_id
		where post_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query,id)

	if err != nil {
		switch{
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
		
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(&c.ID, &c.PostId, &c.UserId, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)

	}

	return comments, nil
}
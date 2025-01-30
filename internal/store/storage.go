package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorNotFound = errors.New("resource not found")
	QueryTimeOutDuration = time.Second*5
	ErrorConflic = errors.New("resource already exists")
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		DeleteByID(context.Context, int64) error 
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata,error)
	}

	Users interface {
		Create(context.Context, *sql.Tx ,*User) error
		GetUserByID(context.Context, int64) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		Activate(context.Context, string) error
		getUserFromInvitation(context.Context,*sql.Tx,string) (*User,error)
		Delete(context.Context,int64 ) error
		
	}

	Comments interface{
		Create(context.Context, *Comment) error
		GetCommentsByPostID(context.Context, int64)([]Comment, error)
	}

	Followers interface{
		Follow(ctx context.Context, followerId, userID int64) error
		Unfollow(ctx context.Context, followerId, userID int64) error
	}
}

func NewPostgresStorage(db *sql.DB) Storage{
	return Storage{
		Posts: &PostsStorage{db},
		Users: &UserStore{db},
		Comments: &CommentsStorage{db},
		Followers: &FollowerStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error{
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

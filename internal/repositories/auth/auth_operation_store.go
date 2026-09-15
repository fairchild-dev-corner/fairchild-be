package repository

import "database/sql"

/**
* AuthRepository
* @Description: Wraps the pooled MySQL connection. All auth persistence
* (users, social accounts, sessions) is implemented as methods on this
* struct across user_repository.go, social_account_repository.go and
* session_repository.go.
**/
type AuthRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{DB: db}
}

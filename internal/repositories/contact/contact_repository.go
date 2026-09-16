package repository

import "database/sql"

/**
* ContactRepository
* @Description: Wraps the pooled MySQL connection. All contact-form
* persistence (contact messages, newsletter subscriptions) is implemented as
* methods on this struct across contact_message_repository.go and
* newsletter_repository.go.
**/
type ContactRepository struct {
	DB *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{DB: db}
}

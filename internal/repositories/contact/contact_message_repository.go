package repository

import (
	models "fairchild_be/internal/models/contact"

	"github.com/gin-gonic/gin"
)

// CreateContactMessage inserts a new contact_messages row and returns its
// generated id.
func (r *ContactRepository) CreateContactMessage(ctx *gin.Context, m *models.ContactMessage) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		`INSERT INTO contact_messages (name, email, message) VALUES (?, ?, ?)`,
		m.Name, m.Email, m.Message,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

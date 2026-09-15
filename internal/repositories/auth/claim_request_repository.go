package repository

import (
	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// CreateClaimRequest inserts a new pending account_claim_requests row and
// returns its generated id. Deliberately does not check whether MemberID
// exists in users - an unmatched member_id is left for staff to notice
// during review rather than rejected up front, so the submission response
// can't be used to enumerate valid member_ids.
func (r *AuthRepository) CreateClaimRequest(ctx *gin.Context, c *models.ClaimRequest) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		`INSERT INTO account_claim_requests (member_id, full_name, proposed_email, proposed_mobile_number, message)
		 VALUES (?, ?, ?, ?, ?)`,
		c.MemberID, c.FullName, c.ProposedEmail, c.ProposedMobileNumber, c.Message,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

package services

import (
	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// SubmitClaimRequestService records a self-submitted request to attach
// contact info to an existing member's account - for members migrated from
// the legacy portal with no email/mobile on file, who therefore can't reach
// Forgot Password. Always succeeds from the caller's point of view (see
// AuthRepository.CreateClaimRequest) - a staff member verifies the member's
// identity out-of-band before approving it, which is what actually attaches
// ProposedEmail/ProposedMobileNumber to the users row (admin dashboard,
// forthcoming - until then, approval is a direct DB update).
func (s *AuthService) SubmitClaimRequestService(ctx *gin.Context, req *models.ClaimAccountRequest) error {
	_, err := s.repo.CreateClaimRequest(ctx, &models.ClaimRequest{
		MemberID:             req.MemberID,
		FullName:             nullIfEmptyStr(req.FullName),
		ProposedEmail:        nullIfEmptyStr(req.ProposedEmail),
		ProposedMobileNumber: nullIfEmptyStr(req.ProposedMobileNumber),
		Message:              nullIfEmptyStr(req.Message),
		Status:               models.ClaimRequestStatusPending,
	})
	return err
}

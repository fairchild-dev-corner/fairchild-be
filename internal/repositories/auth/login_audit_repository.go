package repository

import (
	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// CreateLoginAuditLog persists a record of a badge+password/OTP login attempt.
func (r *AuthRepository) CreateLoginAuditLog(ctx *gin.Context, log *models.LoginAuditLog) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO login_audit_logs (user_id, provider_id, identifier_attempted, ip_address, user_agent, status, reason)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		log.UserID, log.ProviderID, log.IdentifierAttempted, log.IPAddress, log.UserAgent, log.Status, log.Reason,
	)
	return err
}

// CountRecentFailedLogins reports how many failed login attempts were logged
// against the given identifier (member_id) and IP in the trailing
// 15-minute window, computed against MySQL's own clock (consistent with
// created_at being DB-assigned).
func (r *AuthRepository) CountRecentFailedLogins(ctx *gin.Context, identifier, ip string) (identifierFailures, ipFailures int, err error) {
	err = r.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM login_audit_logs
		 WHERE status = 'failed' AND identifier_attempted = ? AND created_at > (NOW() - INTERVAL 15 MINUTE)`,
		identifier,
	).Scan(&identifierFailures)
	if err != nil {
		return 0, 0, err
	}

	err = r.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM login_audit_logs
		 WHERE status = 'failed' AND ip_address = ? AND created_at > (NOW() - INTERVAL 15 MINUTE)`,
		ip,
	).Scan(&ipFailures)
	if err != nil {
		return 0, 0, err
	}

	return identifierFailures, ipFailures, nil
}

// DeleteRecentFailedLogins removes failed login_audit_logs rows within the
// same trailing window CountRecentFailedLogins reads, for identifier or ip -
// development-only, see AuthService.UnlockAccountService. Scoped to that
// window so it never touches older audit history, only the rows actually
// gating the current lockout.
func (r *AuthRepository) DeleteRecentFailedLogins(ctx *gin.Context, identifier, ip string) error {
	_, err := r.DB.ExecContext(ctx,
		`DELETE FROM login_audit_logs
		 WHERE status = 'failed' AND created_at > (NOW() - INTERVAL 15 MINUTE)
		   AND (identifier_attempted = ? OR ip_address = ?)`,
		identifier, ip,
	)
	return err
}

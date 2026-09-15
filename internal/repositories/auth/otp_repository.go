package repository

import (
	"database/sql"
	"errors"

	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// CreateOTPVerification persists a new login OTP challenge and returns its
// generated auto-increment id. expiresInMinutes is applied via MySQL's own
// NOW() rather than binding a Go-computed time.Time, so it stays consistent
// with the DB server's clock even if its session time_zone differs from the
// driver's Loc=UTC setting (see CountRecentFailedLogins for the same
// pattern applied to the lockout window).
func (r *AuthRepository) CreateOTPVerification(ctx *gin.Context, c *models.OTPVerification, expiresInMinutes int) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		`INSERT INTO otp_verifications (reference_id, user_id, member_id, otp_hash, mobile_number, purpose, max_attempts, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, NOW() + INTERVAL ? MINUTE)`,
		c.ReferenceID, c.UserID, c.MemberID, c.OTPHash, c.MobileNumber, c.Purpose, c.MaxAttempts, expiresInMinutes,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// FindActiveOTPVerificationByReference returns the challenge for referenceID
// if it was actually confirmed sent, hasn't been consumed, and hasn't
// expired - or nil if no such challenge exists. A challenge whose SMS failed
// to send is deliberately excluded: nobody could ever legitimately hold its
// code.
func (r *AuthRepository) FindActiveOTPVerificationByReference(ctx *gin.Context, referenceID string) (*models.OTPVerification, error) {
	var c models.OTPVerification

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, reference_id, user_id, member_id, otp_hash, status, send_error, mobile_number, purpose, attempts, max_attempts, expires_at, verified_at, consumed_at, created_at
		 FROM otp_verifications
		 WHERE reference_id = ? AND status = 'sent' AND consumed_at IS NULL AND expires_at > NOW()`,
		referenceID,
	).Scan(
		&c.ID, &c.ReferenceID, &c.UserID, &c.MemberID, &c.OTPHash, &c.Status, &c.SendError, &c.MobileNumber, &c.Purpose,
		&c.Attempts, &c.MaxAttempts, &c.ExpiresAt, &c.VerifiedAt, &c.ConsumedAt, &c.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

// MarkOTPVerified records that the correct code was presented, without yet
// consuming the challenge - see AuthService.VerifyForgotPasswordOTPService,
// which splits "prove you received the code" from "actually reset the
// password" into two separate requests.
func (r *AuthRepository) MarkOTPVerified(ctx *gin.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE otp_verifications SET verified_at = NOW() WHERE id = ? AND consumed_at IS NULL`, id)
	return err
}

// FindVerifiedOTPVerificationByReference returns the challenge for
// referenceID if it was already verified (see MarkOTPVerified) and hasn't
// been consumed yet - or nil otherwise. Deliberately does not re-check
// expires_at: the code itself already proved possession within its window,
// and re-checking it here would race the user's password entry against the
// same short OTP timer instead of letting them take as long as they need
// once verified. The row can still only ever be consumed once, via
// ConsumeOTPVerification.
func (r *AuthRepository) FindVerifiedOTPVerificationByReference(ctx *gin.Context, referenceID string) (*models.OTPVerification, error) {
	var c models.OTPVerification

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, reference_id, user_id, member_id, otp_hash, status, send_error, mobile_number, purpose, attempts, max_attempts, expires_at, verified_at, consumed_at, created_at
		 FROM otp_verifications
		 WHERE reference_id = ? AND verified_at IS NOT NULL AND consumed_at IS NULL`,
		referenceID,
	).Scan(
		&c.ID, &c.ReferenceID, &c.UserID, &c.MemberID, &c.OTPHash, &c.Status, &c.SendError, &c.MobileNumber, &c.Purpose,
		&c.Attempts, &c.MaxAttempts, &c.ExpiresAt, &c.VerifiedAt, &c.ConsumedAt, &c.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

// MarkOTPSent records that the SMS was confirmed delivered by the provider,
// making the challenge eligible for verification.
func (r *AuthRepository) MarkOTPSent(ctx *gin.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE otp_verifications SET status = 'sent' WHERE id = ?`, id)
	return err
}

// MarkOTPSendFailed records that the provider failed to deliver the SMS, so
// the row stays visible for support/debugging instead of looking like an
// ordinary unconsumed challenge nobody ever verified. reason is truncated to
// fit the column - it holds provider error text, never the OTP code itself.
func (r *AuthRepository) MarkOTPSendFailed(ctx *gin.Context, id int64, reason string) error {
	if len(reason) > 255 {
		reason = reason[:255]
	}
	_, err := r.DB.ExecContext(ctx,
		`UPDATE otp_verifications SET status = 'failed', send_error = ? WHERE id = ?`, reason, id)
	return err
}

// IncrementOTPAttempts records one more failed verification attempt and,
// atomically in the same statement, consumes the challenge once attempts
// reaches max_attempts. The CASE evaluates against the row's live attempts
// value under InnoDB's row lock at UPDATE time, not a value read earlier in
// application code, so a burst of concurrent wrong guesses can't each think
// "we're still under the limit" and collectively blow past max_attempts
// uncounted.
//
// The consumed_at assignment is listed BEFORE the attempts assignment
// deliberately: MySQL evaluates a multi-column UPDATE's SET clauses left to
// right, with each later clause seeing any earlier clause's new value for
// that row. Computing consumed_at first means its "attempts + 1" still
// reads the pre-increment value; swapping the order would make it read the
// already-incremented value and consume one attempt too early.
func (r *AuthRepository) IncrementOTPAttempts(ctx *gin.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE otp_verifications
		 SET consumed_at = CASE WHEN attempts + 1 >= max_attempts THEN NOW() ELSE consumed_at END,
		     attempts = attempts + 1
		 WHERE id = ? AND consumed_at IS NULL`, id)
	return err
}

// ConsumeOTPVerification marks the challenge as used, so it can never be
// verified again (success or exhausted-attempts). consumed reports whether
// THIS call actually flipped the row - false means someone else (a
// concurrent request for the same reference_id) already consumed it first,
// which the caller must treat as a failure, not silently proceed past.
func (r *AuthRepository) ConsumeOTPVerification(ctx *gin.Context, id int64) (consumed bool, err error) {
	res, err := r.DB.ExecContext(ctx,
		`UPDATE otp_verifications SET consumed_at = NOW() WHERE id = ? AND consumed_at IS NULL`, id)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

package usecase

import (
	"context"
	"time"
)

type LogoutInput struct {
	RefreshToken string
}

func (u *GoogleFlow) Logout(ctx context.Context, in LogoutInput) error {
	if in.RefreshToken == "" {
		return nil
	}
	h := hashRefresh(u.hashSecret, in.RefreshToken)
	sess, err := u.sessions.GetByRefreshTokenHash(ctx, h)
	if err == nil {
		_ = u.sessions.Revoke(ctx, sess.ID, time.Now())
		audit0(u, ctx, sess.UserID, "auth_logout", map[string]any{})
	}
	return nil
}

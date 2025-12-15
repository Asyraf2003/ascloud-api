package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/ports"
)

type fakeAudit struct{}

func (a fakeAudit) Record(ctx context.Context, e ports.AuditEvent) error { return nil }

type fakeTrust struct{}

func (t fakeTrust) Evaluate(ctx context.Context, s ports.TrustSignals) (ports.TrustDecision, error) {
	return ports.TrustDecision{Allow: true}, nil
}

type memStateStore struct{ v map[string]ports.AuthState }

func memState() *memStateStore { return &memStateStore{v: map[string]ports.AuthState{}} }
func (m *memStateStore) Put(ctx context.Context, k string, s ports.AuthState, ttl time.Duration) error {
	m.v[k] = s
	return nil
}
func (m *memStateStore) GetDel(ctx context.Context, k string) (ports.AuthState, error) {
	s, ok := m.v[k]
	delete(m.v, k)
	if !ok {
		return ports.AuthState{}, context.Canceled
	}
	return s, nil
}

type memSessStore struct{}

func memSess() *memSessStore { return &memSessStore{} }
func (m *memSessStore) Create(ctx context.Context, s domain.Session) (domain.Session, error) {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now()
	return s, nil
}
func (m *memSessStore) GetByID(ctx context.Context, id string) (domain.Session, error) {
	return domain.Session{}, nil
}
func (m *memSessStore) GetByRefreshTokenHash(ctx context.Context, h string) (domain.Session, error) {
	return domain.Session{}, nil
}
func (m *memSessStore) RotateRefreshTokenHash(ctx context.Context, sid, old, nw string, exp time.Time) error {
	return nil
}
func (m *memSessStore) Revoke(ctx context.Context, sid string, at time.Time) error { return nil }

type memIDsRepo struct{ m map[string]uuid.UUID }

func memIDs() *memIDsRepo { return &memIDsRepo{m: map[string]uuid.UUID{}} }
func (r *memIDsRepo) FindAccountIDByIdentity(ctx context.Context, p, sub string) (uuid.UUID, bool, error) {
	id, ok := r.m[p+":"+sub]
	return id, ok, nil
}
func (r *memIDsRepo) UpsertIdentity(ctx context.Context, id uuid.UUID, p, sub, e string, v bool, meta map[string]any) error {
	r.m[p+":"+sub] = id
	return nil
}

type memAccSvc struct{}

func memAcc() memAccSvc { return memAccSvc{} }
func (m memAccSvc) Create(ctx context.Context, in ports.AccountInput) (uuid.UUID, error) {
	return uuid.New(), nil
}

type memTokIssuer struct{}

func memTok() memTokIssuer { return memTokIssuer{} }
func (m memTokIssuer) IssueAccessToken(ctx context.Context, r ports.AccessTokenRequest) (string, time.Time, error) {
	return "jwt.token", time.Now().Add(30 * time.Minute), nil
}

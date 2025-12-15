package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

var ErrStateNotFound = errors.New("auth state not found")

type item struct {
	v   ports.AuthState
	exp time.Time
}

type AuthStateStore struct {
	mu sync.Mutex
	m  map[string]item
}

func NewAuthStateStore() *AuthStateStore {
	return &AuthStateStore{m: map[string]item{}}
}

func (s *AuthStateStore) Put(ctx context.Context, state string, v ports.AuthState, ttl time.Duration) error {
	_ = ctx
	if state == "" {
		return errors.New("state empty")
	}
	if ttl <= 0 {
		return errors.New("ttl invalid")
	}
	s.mu.Lock()
	s.m[state] = item{v: v, exp: time.Now().Add(ttl)}
	s.mu.Unlock()
	return nil
}

func (s *AuthStateStore) GetDel(ctx context.Context, state string) (ports.AuthState, error) {
	_ = ctx
	s.mu.Lock()
	it, ok := s.m[state]
	if ok {
		delete(s.m, state)
	}
	s.mu.Unlock()

	if !ok {
		return ports.AuthState{}, ErrStateNotFound
	}
	if time.Now().After(it.exp) {
		return ports.AuthState{}, ErrStateNotFound
	}
	return it.v, nil
}

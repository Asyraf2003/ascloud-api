package memory

import (
	"context"
	"testing"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

func TestAuthStateStore_GetDel(t *testing.T) {
	s := NewAuthStateStore()
	st := ports.AuthState{Nonce: "n", CodeVerifier: "v", Purpose: "login", CreatedAt: time.Now()}

	if err := s.Put(context.Background(), "k", st, 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetDel(context.Background(), "k"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetDel(context.Background(), "k"); err == nil {
		t.Fatal("expected error on second GetDel")
	}
}

func TestAuthStateStore_Expired(t *testing.T) {
	s := NewAuthStateStore()
	st := ports.AuthState{Nonce: "n", CodeVerifier: "v", Purpose: "login", CreatedAt: time.Now()}

	if err := s.Put(context.Background(), "k", st, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := s.GetDel(context.Background(), "k"); err == nil {
		t.Fatal("expected expired state error")
	}
}

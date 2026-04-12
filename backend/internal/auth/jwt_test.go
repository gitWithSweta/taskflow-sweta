package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSignAndParseToken(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-signing-only")
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	raw, err := SignToken(secret, userID, email, sessionID, expiresAt)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseToken(secret, raw)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Email != email {
		t.Fatalf("email: got %q want %q", claims.Email, email)
	}
	if claims.UserID != userID.String() {
		t.Fatalf("user_id: got %q want %q", claims.UserID, userID.String())
	}
	sid, err := claims.SessionID()
	if err != nil {
		t.Fatalf("session_id parse: %v", err)
	}
	if sid != sessionID {
		t.Fatalf("session_id: got %v want %v", sid, sessionID)
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	secret := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	raw, err := SignToken(secret, uuid.New(), "a@b.co", uuid.New(), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseToken([]byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), raw)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestParseExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-signing-only")

	raw, err := SignToken(secret, uuid.New(), "x@y.co", uuid.New(), time.Now().Add(-time.Second))
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseToken(secret, raw)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestPasswordBcryptCost(t *testing.T) {
	h, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(h, "correct-horse-battery-staple") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(h, "wrong") {
		t.Fatal("expected mismatch")
	}
}

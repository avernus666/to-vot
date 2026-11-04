package security

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyPassword("correct-horse-battery-staple", hash)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
	ok, err = VerifyPassword("not-the-password", hash)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected password verification success")
	}
}

func TestHashPasswordRejectsShortPassword(t *testing.T) {
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("expected short password to be rejected")
	}
}

package cmd

import "testing"

func TestValidateTLS(t *testing.T) {
	if err := validateTLS("", ""); err != nil {
		t.Fatalf("both empty should be ok, got %v", err)
	}
	if err := validateTLS("cert", "key"); err != nil {
		t.Fatalf("both set should be ok, got %v", err)
	}
	if err := validateTLS("cert", ""); err == nil {
		t.Fatal("cert without key should fail")
	}
	if err := validateTLS("", "key"); err == nil {
		t.Fatal("key without cert should fail")
	}
}

func TestRootCmdBsPortZero(t *testing.T) {
	oldBsPort := bsPort
	bsPort = 0
	defer func() { bsPort = oldBsPort }()
	if err := rootCmd(); err == nil || err.Error() != "battle server port must be specified and non-zero" {
		t.Fatalf("expected bsPort error, got %v", err)
	}
}

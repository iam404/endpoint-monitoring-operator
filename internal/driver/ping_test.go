package driver

import "testing"

func TestNewPingDriver(t *testing.T) {
	if _, err := NewPingDriver(""); err == nil {
		t.Error("expected error for empty endpoint")
	}

	d, err := NewPingDriver("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "ping" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "ping")
	}
	if d.GetEndpoint() != "example.com" {
		t.Errorf("GetEndpoint() = %q", d.GetEndpoint())
	}
}

func TestPingDriver_Check_Failure(t *testing.T) {
	// The ping driver dials <endpoint>:80. An unresolvable host fails fast and
	// deterministically without requiring network access.
	d, err := NewPingDriver("no-such-host.invalid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := d.Check()
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if result.Success {
		t.Error("expected Success = false for an unresolvable host")
	}
	if result.Error == nil {
		t.Error("expected result.Error to be set on dial failure")
	}
}

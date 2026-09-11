package driver

import "testing"

func TestNewDNSDriver(t *testing.T) {
	if _, err := NewDNSDriver(""); err == nil {
		t.Error("expected error for empty endpoint")
	}

	d, err := NewDNSDriver("localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "dns" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "dns")
	}
	if d.GetEndpoint() != "localhost" {
		t.Errorf("GetEndpoint() = %q", d.GetEndpoint())
	}
}

func TestDNSDriver_Check_Success(t *testing.T) {
	// "localhost" resolves from the hosts file without needing a network.
	d, err := NewDNSDriver("localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := d.Check()
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success = true resolving localhost, got false (message: %s)", result.Message)
	}
}

func TestDNSDriver_Check_Failure(t *testing.T) {
	// The .invalid TLD is reserved (RFC 2606) and guaranteed never to resolve.
	d, err := NewDNSDriver("no-such-host.invalid")
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
		t.Error("expected result.Error to be set on resolution failure")
	}
}

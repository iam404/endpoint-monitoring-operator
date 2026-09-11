package driver

import (
	"net"
	"testing"
)

func TestNewTCPDriver(t *testing.T) {
	if _, err := NewTCPDriver(""); err == nil {
		t.Error("expected error for empty endpoint")
	}

	d, err := NewTCPDriver("127.0.0.1:1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "tcp" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "tcp")
	}
	if d.GetEndpoint() != "127.0.0.1:1234" {
		t.Errorf("GetEndpoint() = %q", d.GetEndpoint())
	}
}

func TestTCPDriver_Check_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	// Accept and immediately close connections so dials succeed.
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	d, err := NewTCPDriver(ln.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := d.Check()
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success = true, got false (message: %s)", result.Message)
	}
}

func TestTCPDriver_Check_Failure(t *testing.T) {
	// Bind a port then close it so nothing is listening there.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	d, err := NewTCPDriver(addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := d.Check()
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if result.Success {
		t.Error("expected Success = false when nothing is listening")
	}
	if result.Error == nil {
		t.Error("expected result.Error to be set on connection failure")
	}
}

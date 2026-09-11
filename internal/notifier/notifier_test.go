package notifier

import "testing"

func TestShouldAlert(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		status  string
		want    bool
	}{
		{"empty list defaults to failure only", nil, "failure", true},
		{"empty list skips success", nil, "success", false},
		{"status in list", []string{"success", "failure"}, "success", true},
		{"status not in list", []string{"failure"}, "success", false},
		{"failure in list", []string{"failure"}, "failure", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldAlert(tt.allowed, tt.status); got != tt.want {
				t.Errorf("ShouldAlert(%v, %q) = %v, want %v", tt.allowed, tt.status, got, tt.want)
			}
		})
	}
}

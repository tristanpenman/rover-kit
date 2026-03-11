package main

import (
	"strings"
	"testing"

	"rover-kit/pkg/common"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		expected    any
		expectedErr string
	}{
		{
			name:     "forwards command",
			payload:  []byte(`{"type":"forwards"}`),
			expected: common.ForwardsCommand{Type: common.CommandForwards},
		},
		{
			name:     "backwards command",
			payload:  []byte(`{"type":"backwards"}`),
			expected: common.BackwardsCommand{Type: common.CommandBackwards},
		},
		{
			name:     "spin clockwise command",
			payload:  []byte(`{"type":"spin_cw"}`),
			expected: common.SpinCWCommand{Type: common.CommandSpinCW},
		},
		{
			name:     "spin counterclockwise command",
			payload:  []byte(`{"type":"spin_ccw"}`),
			expected: common.SpinCCWCommand{Type: common.CommandSpinCCW},
		},
		{
			name:     "stop command",
			payload:  []byte(`{"type":"stop"}`),
			expected: common.StopCommand{Type: common.CommandStop},
		},
		{
			name:     "throttle active",
			payload:  []byte(`{"type":"throttle","value":0.3}`),
			expected: throttleResponse{Type: common.CommandThrottle, Active: true},
		},
		{
			name:     "throttle inactive",
			payload:  []byte(`{"type":"throttle","value":0}`),
			expected: throttleResponse{Type: common.CommandThrottle, Active: false},
		},
		{
			name:        "invalid type",
			payload:     []byte(`{"type":"unsupported"}`),
			expectedErr: "invalid payload type",
		},
		{
			name:        "invalid json",
			payload:     []byte(`{"type":`),
			expectedErr: "bad request:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommand(tt.payload)
			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if !contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("unexpected result: got %#v, want %#v", result, tt.expected)
			}
		})
	}
}

func contains(value, expectedSubstring string) bool {
	return strings.Contains(value, expectedSubstring)
}

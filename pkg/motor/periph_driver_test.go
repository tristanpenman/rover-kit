package motor

import "testing"

func TestPWMDutyFromThrottle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		throttle float64
		want     int
	}{
		{name: "full reverse is clamped", throttle: -2.0, want: periphPWMMax},
		{name: "full reverse", throttle: -1.0, want: periphPWMMax},
		{name: "stopped", throttle: 0.0, want: 0},
		{name: "half forward", throttle: 0.5, want: 2048},
		{name: "full forward", throttle: 1.0, want: periphPWMMax},
		{name: "full forward is clamped", throttle: 2.0, want: periphPWMMax},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := int(pwmDutyFromThrottle(tc.throttle)); got != tc.want {
				t.Fatalf("pwmDutyFromThrottle(%v) = %d, want %d", tc.throttle, got, tc.want)
			}
		})
	}
}

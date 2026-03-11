package common

import "testing"

func TestEnvOrDefault(t *testing.T) {
	const key = "ROVER_KIT_TEST_ENV"
	t.Setenv(key, "configured")

	value := EnvOrDefault(key, "fallback")
	if value != "configured" {
		t.Fatalf("expected configured value, got %q", value)
	}
}

func TestEnvOrDefaultFallsBack(t *testing.T) {
	const key = "ROVER_KIT_TEST_ENV_EMPTY"
	t.Setenv(key, "")

	value := EnvOrDefault(key, "fallback")
	if value != "fallback" {
		t.Fatalf("expected fallback value, got %q", value)
	}
}

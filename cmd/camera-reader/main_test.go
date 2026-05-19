package main

import (
	"os"
	"testing"
	"time"

	"rover-kit/pkg/camera"
)

func TestCreateProviderDummy(t *testing.T) {
	t.Setenv("CAMERA_INTERVAL_MS", "25")

	provider, err := createProvider("dummy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dummy, ok := provider.(*camera.DummyProvider)
	if !ok {
		t.Fatalf("expected dummy provider, got %T", provider)
	}
	if dummy.Interval != 25*time.Millisecond {
		t.Fatalf("unexpected interval: got %s", dummy.Interval)
	}
}

func TestCreateProviderLibcamera(t *testing.T) {
	t.Setenv("CAMERA_INTERVAL_MS", "50")
	t.Setenv("CAMERA_CAPTURE_TIMEOUT_MS", "100")
	t.Setenv("CAMERA_WIDTH", "640")
	t.Setenv("CAMERA_HEIGHT", "480")
	t.Setenv("LIBCAMERA_STILL_PATH", "/usr/bin/libcamera-still")

	provider, err := createProvider("libcamera")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	libcamera, ok := provider.(*camera.LibcameraProvider)
	if !ok {
		t.Fatalf("expected libcamera provider, got %T", provider)
	}
	if libcamera.Interval != 50*time.Millisecond {
		t.Fatalf("unexpected interval: got %s", libcamera.Interval)
	}
	if libcamera.Timeout != 100*time.Millisecond {
		t.Fatalf("unexpected timeout: got %s", libcamera.Timeout)
	}
	if libcamera.Width != 640 || libcamera.Height != 480 {
		t.Fatalf("unexpected size: got %dx%d", libcamera.Width, libcamera.Height)
	}
	if libcamera.CommandPath != "/usr/bin/libcamera-still" {
		t.Fatalf("unexpected command path: %s", libcamera.CommandPath)
	}
}

func TestCreateProviderUnsupported(t *testing.T) {
	_, err := createProvider("nope")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDurationFromEnvMSRejectsZero(t *testing.T) {
	t.Setenv("CAMERA_INTERVAL_MS", "0")

	_, err := durationFromEnvMS("CAMERA_INTERVAL_MS", 1000)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIntFromEnvRejectsInvalidValue(t *testing.T) {
	t.Setenv("CAMERA_WIDTH", "wide")

	_, err := intFromEnv("CAMERA_WIDTH", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIntFromEnvUsesDefault(t *testing.T) {
	if err := os.Unsetenv("CAMERA_WIDTH"); err != nil {
		t.Fatalf("unset env: %v", err)
	}

	value, err := intFromEnv("CAMERA_WIDTH", 320)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 320 {
		t.Fatalf("unexpected value: got %d", value)
	}
}

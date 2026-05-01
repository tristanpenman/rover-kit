package common

import "testing"

func TestLineBuffer(t *testing.T) {
	var lb LineBuffer

	lines := lb.Append([]byte("hello"))
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(lines))
	}

	lines = lb.Append([]byte(" world\ntest"))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "hello world" {
		t.Fatalf("expected hello world, got %s", lines[0])
	}

	lines = lb.Append([]byte(" 123\r\nfoo"))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "test 123" {
		t.Fatalf("expected test 123, got %s", lines[0])
	}

	lines = lb.Append([]byte(" bar\r\nfoo baz\n"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "foo bar" {
		t.Fatalf("expected foo bar, got %s", lines[0])
	}
	if lines[1] != "foo baz" {
		t.Fatalf("expected foo baz, got %s", lines[1])
	}

	lines = lb.Append([]byte("\r\n"))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "" {
		t.Fatalf("expected empty string, got %s", lines[0])
	}
}

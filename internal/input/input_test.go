package input

import (
	"strings"
	"testing"
)

func TestResolvePrefersArgsOverStdin(t *testing.T) {
	t.Parallel()

	got, err := Resolve([]string{"hello", "world"}, strings.NewReader("ignored"), true)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if got != "hello world" {
		t.Fatalf("Resolve() = %q, want %q", got, "hello world")
	}
}

func TestResolveReadsStdinWhenNoArgs(t *testing.T) {
	t.Parallel()

	got, err := Resolve(nil, strings.NewReader("  from stdin  \n"), true)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if got != "from stdin" {
		t.Fatalf("Resolve() = %q, want %q", got, "from stdin")
	}
}

func TestResolveRequiresInput(t *testing.T) {
	t.Parallel()

	_, err := Resolve(nil, strings.NewReader(""), false)
	if err == nil {
		t.Fatal("Resolve() expected an error, got nil")
	}
	if err != ErrNoInput {
		t.Fatalf("Resolve() error = %v, want %v", err, ErrNoInput)
	}
}

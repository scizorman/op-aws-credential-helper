package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCommandOTPSource_TrimsTrailingNewline(t *testing.T) {
	source := &commandOTPSource{command: "echo 123456"}

	otp, err := source.OTP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if otp != "123456" {
		t.Errorf("otp = %q, want %q", otp, "123456")
	}
}

func TestCommandOTPSource_TrimsSurroundingWhitespace(t *testing.T) {
	source := &commandOTPSource{command: `printf '  123456\n'`}

	otp, err := source.OTP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if otp != "123456" {
		t.Errorf("otp = %q, want %q", otp, "123456")
	}
}

func TestCommandOTPSource_DoesNotValidateDigitCount(t *testing.T) {
	source := &commandOTPSource{command: "echo 12345678"}

	otp, err := source.OTP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if otp != "12345678" {
		t.Errorf("otp = %q, want %q", otp, "12345678")
	}
}

func TestCommandOTPSource_EmptyOutputFailsFast(t *testing.T) {
	source := &commandOTPSource{command: "true"}

	_, err := source.OTP(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "true") {
		t.Errorf("error = %q, want it to contain the command %q", err.Error(), "true")
	}
}

func TestCommandOTPSource_MultiLineOutputFailsFast(t *testing.T) {
	source := &commandOTPSource{command: `printf 'warn\n123456\n'`}

	_, err := source.OTP(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), source.command) {
		t.Errorf("error = %q, want it to contain the command %q", err.Error(), source.command)
	}
}

func TestCommandOTPSource_StderrIsPassedThrough(t *testing.T) {
	var stderr bytes.Buffer
	source := &commandOTPSource{command: "echo touch-your-key >&2; echo 123456", stderr: &stderr}

	otp, err := source.OTP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if otp != "123456" {
		t.Errorf("otp = %q, want %q", otp, "123456")
	}
	if !strings.Contains(stderr.String(), "touch-your-key") {
		t.Errorf("stderr = %q, want it to contain %q", stderr.String(), "touch-your-key")
	}
}

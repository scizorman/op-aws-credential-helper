package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type OTPSource interface {
	OTP(ctx context.Context) (string, error)
}

type ttyOTPSource struct{}

func (s *ttyOTPSource) OTP(ctx context.Context) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tty.Close()
	}()

	if _, err := fmt.Fprint(tty, "Enter MFA code: "); err != nil {
		return "", err
	}
	var code string
	if _, err := fmt.Fscanln(tty, &code); err != nil {
		return "", err
	}
	return code, nil
}

// commandOTPSource runs an external command via /bin/sh -c to obtain an MFA
// code, matching aws-vault's mfa_process. It performs no argument splitting
// of its own: the shell owns quoting and piping.
type commandOTPSource struct {
	command string
	stderr  io.Writer
}

func (s *commandOTPSource) OTP(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", s.command)

	stderr := s.stderr
	if stderr == nil {
		// The credential_process caller captures our own stderr, so an MFA
		// provider's touch prompt (e.g. ykman) needs the terminal directly.
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			stderr = os.Stderr
		} else {
			defer func() {
				_ = tty.Close()
			}()
			cmd.Stdin = tty
			stderr = tty
		}
	}
	cmd.Stderr = stderr

	// cmd.Output() requires cmd.Stdout to be unset; cmd.Stderr may be set
	// beforehand, which only means ExitError.Stderr stays empty since the
	// output already streamed to stderr above.
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mfa-process command %s failed: %w", s.command, err)
	}

	otp := strings.TrimSpace(string(out))
	if len(strings.Fields(otp)) != 1 {
		return "", fmt.Errorf("mfa-process command %s produced invalid output %q", s.command, otp)
	}

	return otp, nil
}

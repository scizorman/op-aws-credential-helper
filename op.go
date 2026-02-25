package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type opCLICredentialSource struct {
	cliPath string
	OpAwsItem
}

func (s *opCLICredentialSource) Retrieve(ctx context.Context) (aws.Credentials, error) {
	fields := fmt.Sprintf("label=%s,label=%s", s.AccessKeyIDField, s.SecretAccessKeyField)
	cmd := exec.CommandContext(ctx, s.cliPath,
		"item", "get", s.Item,
		"--vault", s.Vault,
		"--fields", fields,
		"--format", "json",
	)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return aws.Credentials{}, fmt.Errorf("failed to get op item: %w\n%s", err, exitErr.Stderr)
		}
		return aws.Credentials{}, err
	}

	var items []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return aws.Credentials{}, err
	}

	var creds aws.Credentials
	for _, item := range items {
		switch item.Label {
		case s.AccessKeyIDField:
			creds.AccessKeyID = item.Value
		case s.SecretAccessKeyField:
			creds.SecretAccessKey = item.Value
		}
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return aws.Credentials{}, fmt.Errorf("missing credentials in op output")
	}
	return creds, nil
}

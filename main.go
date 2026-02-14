package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/processcreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var version = "dev"

var cli struct {
	Profile                string           `default:"default" help:"AWS config profile name."`
	Duration               time.Duration    `default:"12h" help:"STS session duration."`
	OpVault                string           `required:"" help:"1Password vault name."`
	OpItem                 string           `required:"" help:"1Password item name."`
	OpAccessKeyIDField     string           `default:"username" help:"1Password field name for access key ID." name:"op-access-key-id-field"`
	OpSecretAccessKeyField string           `default:"credential" help:"1Password field name for secret access key." name:"op-secret-access-key-field"`
	OpCLIPath              string           `default:"op" help:"Path to 1Password CLI." name:"op-cli-path"`
	Version                kong.VersionFlag `help:"Show version."`
}

func main() {
	kong.Parse(&cli,
		kong.Name("op-aws-credential-helper"),
		kong.Description("AWS credential_process helper that retrieves credentials from 1Password with MFA session caching"),
		kong.Vars{"version": version},
	)

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cached, err := readCache(cli.Profile)
	if err == nil && time.Now().Add(5*time.Minute).Before(*cached.Expiration) {
		return json.NewEncoder(os.Stdout).Encode(cached)
	}

	cfg, err := config.LoadSharedConfigProfile(ctx, cli.Profile)
	if err != nil {
		return err
	}

	credSource := &opCLICredentialSource{
		cliPath:              cli.OpCLIPath,
		vault:                cli.OpVault,
		item:                 cli.OpItem,
		accessKeyIDField:     cli.OpAccessKeyIDField,
		secretAccessKeyField: cli.OpSecretAccessKeyField,
	}
	creds, err := credSource.Retrieve(ctx)
	if err != nil {
		return err
	}

	otpSource := &ttyOTPSource{}
	otp, err := otpSource.OTP(ctx)
	if err != nil {
		return err
	}

	stsClient := sts.New(sts.Options{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, ""),
	})
	out, err := stsClient.GetSessionToken(ctx, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(int32(cli.Duration.Seconds())),
		SerialNumber:    aws.String(cfg.MFASerial),
		TokenCode:       aws.String(otp),
	})
	if err != nil {
		return err
	}

	resp := processcreds.CredentialProcessResponse{
		Version:         1,
		AccessKeyID:     aws.ToString(out.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(out.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(out.Credentials.SessionToken),
		Expiration:      out.Credentials.Expiration,
	}
	if err := writeCache(cli.Profile, resp); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(resp)
}

func cachePath(profile string) (string, error) {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(home, ".cache")
	}
	hash := sha1.Sum([]byte(profile))
	n := fmt.Sprintf("%x.json", hash)
	return filepath.Join(cacheDir, "op-aws-credential-helper", n), nil
}

func readCache(profile string) (processcreds.CredentialProcessResponse, error) {
	path, err := cachePath(profile)
	if err != nil {
		return processcreds.CredentialProcessResponse{}, err
	}
	d, err := os.ReadFile(path)
	if err != nil {
		return processcreds.CredentialProcessResponse{}, err
	}
	var resp processcreds.CredentialProcessResponse
	if err := json.Unmarshal(d, &resp); err != nil {
		return processcreds.CredentialProcessResponse{}, err
	}
	return resp, nil
}

func writeCache(profile string, resp processcreds.CredentialProcessResponse) error {
	path, err := cachePath(profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	d, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return os.WriteFile(path, d, 0600)
}

type GetSessionTokenAPIClient interface {
	GetSessionToken(ctx context.Context, param *sts.GetSessionTokenInput, optFns ...func(*sts.Options)) (*sts.GetSessionTokenOutput, error)
}

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

type opCLICredentialSource struct {
	cliPath              string
	vault                string
	item                 string
	accessKeyIDField     string
	secretAccessKeyField string
}

func (s *opCLICredentialSource) Retrieve(ctx context.Context) (aws.Credentials, error) {
	fields := fmt.Sprintf("label=%s,label=%s", s.accessKeyIDField, s.secretAccessKeyField)
	cmd := exec.CommandContext(ctx, s.cliPath,
		"item", "get", s.item,
		"--vault", s.vault,
		"--fields", fields,
		"--format", "json",
	)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
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
		case s.accessKeyIDField:
			creds.AccessKeyID = item.Value
		case s.secretAccessKeyField:
			creds.SecretAccessKey = item.Value
		}
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return aws.Credentials{}, fmt.Errorf("missing credentials in op output")
	}
	return creds, nil
}

# op-aws-credential-helper

AWS credential_process helper that retrieves credentials from 1Password with MFA session caching.

This tool retrieves IAM credentials stored in 1Password, performs MFA authentication, and obtains temporary credentials via AWS STS.
It implements the `credential_process` protocol, making it compatible with not only the AWS CLI but also Terraform, boto3, and any other tool that uses the AWS SDK.
Temporary credentials are cached per profile and reused until expiration.

## Requirements

- **Unix-like OS** (Linux, macOS) — Uses `/dev/tty` for MFA input
- **1Password CLI (`op`)** — Used to retrieve credentials
- **AWS Account** — Requires an IAM user with an MFA device

## Installation

### Download from GitHub Release

Download the binary for your platform from the latest release.

https://github.com/scizorman/op-aws-credential-helper/releases

### go install

Requires Go 1.25+.

```bash
go install github.com/scizorman/op-aws-credential-helper@latest
```

## Setup

### 1Password

Store your AWS credentials in 1Password.

By default, the tool expects the Access Key ID in the `username` field and the Secret Access Key in the `credential` field.
Field names can be customized via command-line options.

### AWS CLI

Configure `~/.aws/config` as follows:

```ini
[profile example]
region = ap-northeast-1
mfa_serial = arn:aws:iam::123456789012:mfa/user
credential_process = op-aws-credential-helper --op-vault <vault> --op-item <item>
```

`mfa_serial` is the ARN of the MFA device assigned to your IAM user.
`credential_process` specifies the command line for op-aws-credential-helper.

#### WSL

On WSL, you can use the Windows-side 1Password CLI by specifying the path with `--op-cli-path`:

```ini
[profile example]
region = ap-northeast-1
mfa_serial = arn:aws:iam::123456789012:mfa/user
credential_process = op-aws-credential-helper --op-vault <vault> --op-item <item> --op-cli-path /mnt/c/Program\ Files/1Password\ CLI/op.exe
```

This allows you to leverage Windows Hello biometric authentication from WSL.

## Usage

### CLI Options

| Flag | Default | Required | Description |
|------|---------|----------|-------------|
| `--profile` | `default` | No | AWS config profile name |
| `--duration` | `12h` | No | STS session duration |
| `--op-vault` | - | Yes | 1Password vault name |
| `--op-item` | - | Yes | 1Password item name |
| `--op-access-key-id-field` | `username` | No | Field name for Access Key ID |
| `--op-secret-access-key-field` | `credential` | No | Field name for Secret Access Key |
| `--op-cli-path` | `op` | No | Path to 1Password CLI |

## How it works

1. Load region and mfa_serial from `~/.aws/config`
2. Retrieve IAM access keys using 1Password CLI (`op item get`)
3. Prompt for MFA code via `/dev/tty`
4. Obtain temporary credentials via STS `GetSessionToken`
5. Cache temporary credentials and output credential_process JSON to stdout

### Cache

Temporary credentials are cached per profile.

- **Cache location**: `$XDG_CACHE_HOME/op-aws-credential-helper/<profile>.json` (defaults to `~/.cache/op-aws-credential-helper/<profile>.json`)
- **Cache granularity**: One file per profile
- **Cache invalidation**:
  - When configuration parameters (vault, item, mfa_serial, field names) change
  - 5 minutes before expiration

## Comparison

| Aspect | aws-vault | 1Password Shell Plugin | op-aws-credential-helper |
|--------|-----------|----------------------|--------------------------|
| Credential Storage | OS keystore | 1Password | 1Password |
| Injection Method | credential_process / env vars | env vars | credential_process |
| Tool Support | All credential_process tools | Plugin-supported commands only | All credential_process tools |
| MFA Token | Interactive prompt / mfa_process | 1Password TOTP auto-retrieval | Interactive input via /dev/tty |
| External Dependencies | None (single binary) | 1Password Desktop App | op CLI |

This tool provides `credential_process` support for credentials stored in 1Password, making them available to all AWS SDK-compatible tools.

## License

MIT License

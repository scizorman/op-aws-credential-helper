# op-aws-credential-process

AWS credential_process implementation that retrieves credentials from 1Password with MFA session caching.

This tool retrieves IAM credentials stored in 1Password, performs MFA authentication, and obtains temporary credentials via AWS STS.
It implements the `credential_process` protocol, making it compatible with not only the AWS CLI but also Terraform, boto3, and any other tool that uses the AWS SDK.
Temporary credentials are cached per profile and reused until expiration.

## Requirements

- **Unix-like OS** (Linux, macOS) — Uses `/dev/tty` for MFA input
- **1Password CLI (`op`) v2** — Used to retrieve credentials
- **AWS Account** — Requires an IAM user with an MFA device

## Installation

### Download from GitHub Release

Download the binary for your platform from the latest release.

https://github.com/scizorman/op-aws-credential-process/releases

### go install

Requires Go 1.25+.

```bash
go install github.com/scizorman/op-aws-credential-process@latest
```

## Setup

### 1Password

#### Setting up the op CLI

If desktop app integration is enabled, the `op` CLI will unlock via biometric authentication automatically, requiring no manual sign-in.
If integration is disabled, you must sign in manually with `eval $(op signin)`.

#### Storing AWS credentials

Store your AWS credentials in 1Password.

By default, the tool expects the Access Key ID in the `Access key ID` field and the Secret Access Key in the `Secret access key` field.
Field names can be customized via `--op-access-key-id-field` and `--op-secret-access-key-field` flags.

### AWS CLI

Configure `~/.aws/config` as follows:

```ini
[profile example]
region = ap-northeast-1
mfa_serial = arn:aws:iam::123456789012:mfa/user
credential_process = op-aws-credential-process --op-vault <vault> --op-item <item>
```

`mfa_serial` is the ARN of the MFA device assigned to your IAM user.
`credential_process` specifies the command line for op-aws-credential-process.

#### WSL

On WSL, you can use the Windows-side 1Password CLI by specifying the path with `--op-cli-path`:

```ini
[profile example]
region = ap-northeast-1
mfa_serial = arn:aws:iam::123456789012:mfa/user
credential_process = op-aws-credential-process --op-vault <vault> --op-item <item> --op-cli-path /mnt/c/Program\ Files/1Password\ CLI/op.exe
```

This allows you to leverage Windows Hello biometric authentication from WSL.

#### Cross-account access with AssumeRole

Currently, this tool performs `GetSessionToken` to obtain MFA-authenticated temporary credentials.
For cross-account access, combine this with AWS CLI's `source_profile` and `role_arn` settings.

**Example configuration**:

```ini
# Base profile using credential_process
[profile base]
region = ap-northeast-1
mfa_serial = arn:aws:iam::111111111111:mfa/user
credential_process = op-aws-credential-process --op-vault <vault> --op-item <item>

# Cross-account profile using AssumeRole
[profile cross-account]
region = ap-northeast-1
source_profile = base
role_arn = arn:aws:iam::222222222222:role/CrossAccountRole
```

The AWS CLI will first retrieve temporary credentials from the `base` profile, then use them to assume the role specified in `role_arn`.

## Usage

### CLI Options

| Flag | Default | Required | Description |
|------|---------|----------|-------------|
| `--profile` | `default` | No | AWS config profile name |
| `--duration` | `12h` | No | STS session duration |
| `--op-vault` | - | Yes | 1Password vault name |
| `--op-item` | - | Yes | 1Password item name |
| `--op-access-key-id-field` | `Access key ID` | No | Field name for Access Key ID |
| `--op-secret-access-key-field` | `Secret access key` | No | Field name for Secret Access Key |
| `--op-cli-path` | `op` | No | Path to 1Password CLI |

### Cache

Temporary credentials are cached at `$XDG_CACHE_HOME/op-aws-credential-process/<profile>.json` (defaults to `~/.cache/op-aws-credential-process/<profile>.json`).

## Comparison

| Aspect | aws-vault | 1Password Shell Plugin | op-aws-credential-process |
|--------|-----------|----------------------|---------------------------|
| Credential Storage | OS keystore | 1Password | 1Password |
| Injection Method | credential_process / env vars | env vars | credential_process |
| Tool Support | All credential_process tools | Plugin-supported commands only | All credential_process tools |
| MFA Token | Interactive prompt / mfa_process | 1Password TOTP auto-retrieval | Interactive input via /dev/tty |
| External Dependencies | None (single binary) | 1Password Desktop App | op CLI |

## License

MIT License

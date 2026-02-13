package main

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"
)

var version = "dev"

var cli struct {
	Profile   string           `arg:"" help:"AWS config profile name."`
	Duration  time.Duration    `default:"12h" help:"STS session duration."`
	OpCLIPath string           `default:"op" help:"Path to 1Password CLI." name:"op-cli-path"`
	Version   kong.VersionFlag `help:"Show version."`
}

func main() {
	kong.Parse(
		&cli,
		kong.Name("op-aws-credential-helper"),
		kong.Description("AWS credential_process helper that retrieves credentials from 1Password with MFA session caching"),
		kong.Vars{"version": version},
	)

	fmt.Printf("profile=%s duration=%s op-cli-path=%s\n", cli.Profile, cli.Duration, cli.OpCLIPath)
}

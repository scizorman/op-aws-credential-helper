{
  description = "AWS credential_process implementation that retrieves credentials from 1Password with MFA session caching";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
    }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
    in
    {
      packages = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.buildGoModule rec {
          pname = "op-aws-credential-process";
          version = "0.1.1";
          src = ./.;
          vendorHash = "sha256-UF0NkoWKLoODdcq+mwgcFatEaLeF+ee+wa+/dwot2RM=";
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];
        };
      });

      devShells = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.mkShell {
          packages = with nixpkgs.legacyPackages.${system}; [
            go_1_26
            gopls
            delve
            golangci-lint
            goreleaser
            nixfmt
          ];
        };
      });
    };
}

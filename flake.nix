{
  description = "AWS credential_process helper that retrieves credentials from 1Password with MFA session caching";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      treefmt-nix,
    }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      treefmtEval = forAllSystems (
        system:
        treefmt-nix.lib.evalModule nixpkgs.legacyPackages.${system} {
          projectRootFile = "flake.nix";
          programs.gofmt.enable = true;
          programs.nixfmt.enable = true;
        }
      );
    in
    {
      packages = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.buildGoModule {
          pname = "op-aws-credential-helper";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-UF0NkoWKLoODdcq+mwgcFatEaLeF+ee+wa+/dwot2RM=";
        };
      });

      devShells = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.mkShell {
          packages = with nixpkgs.legacyPackages.${system}; [
            go
            golangci-lint
          ];
        };
      });

      checks = forAllSystems (system: {
        formatting = treefmtEval.${system}.config.build.check self;
        package = self.packages.${system}.default;
      });

      formatter = forAllSystems (system: treefmtEval.${system}.config.build.wrapper);
    };
}

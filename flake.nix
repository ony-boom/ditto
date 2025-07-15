{
  description = "Ditto, manage arch package declaratively";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages = {
          default = pkgs.buildGoModule {
            src = self;
            pname = "ditto";
            version = "0.1.0";
            vendorHash = "sha256-AM4ZdnMRSBDEqqMDFtIkMAkaaSWkaEjZdV6mFsSH9y8=";
          };
        };

        overlays.default = final: prev: {
          ditto = self.packages.${prev.system}.default;
        };
      }
    );
}

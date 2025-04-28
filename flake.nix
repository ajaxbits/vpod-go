{
  description = "vpod -- beware the YouTube to podcast feed pipeline";
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    yt-dlp.url = "github:NixOS/nixpkgs/master";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    # https://flake.parts/module-arguments.html
    flake-parts.lib.mkFlake { inherit inputs; } (top: {
      flake = {
        nixosModules.vpod = import ./nix/tests/server-module.nix {
          inherit (self.x86_64-linux.packages) vpod;
        };

        nixConfig = {
          extra-substituters = [
            "https://cache.garnix.io"
          ];
          extra-trusted-public-keys = [
            "cache.garnix.io:CTFPyKSLcx5RMJKfLo5EEPUObbA78b0YQ2DTCJXqr9g="
          ];
        };
      };
      systems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      perSystem =
        {
          pkgs,
          self',
          system,
          ...
        }:
        let
          inherit (import inputs.yt-dlp { inherit system; }) yt-dlp;
          inherit (pkgs) sqlite;

          name = "vpod";
          runtimeDeps = [
            sqlite
            yt-dlp
          ];
        in
        {
          packages =
            let
              lastModifiedDate = top.self.lastModifiedDate or top.self.lastModified or "19700101";
              version = builtins.substring 0 8 lastModifiedDate;
            in
            {
              ${name} = pkgs.callPackage ./nix/package.nix {
                inherit (pkgs) buildGoModule makeWrapper lib;
                inherit
                  name
                  pkgs
                  runtimeDeps
                  version
                  ;
              };
              oci-image = pkgs.callPackage ./nix/oci-image.nix {
                package = self'.packages.${name};
                inherit (pkgs) dockerTools lib;
              };
              default = self'.packages.${name};
            };

          devShells = {
            ${name} = pkgs.callPackage ./nix/shell.nix { inherit pkgs runtimeDeps; };
            default = self'.devShells.${name};
          };

          checks = import ./nix/tests { self = self'; };
        };
    });
}

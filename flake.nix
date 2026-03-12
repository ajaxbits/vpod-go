{
  description = "vpod -- beware the YouTube to podcast feed pipeline";
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    yt-dlp.url = "github:ajaxbits/yt-dlp-flake";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    inputs@{ flake-parts, ... }:
    # https://flake.parts/module-arguments.html
    flake-parts.lib.mkFlake { inherit inputs; } (
      top@{ withSystem, ... }:
      {
        flake = {
          nixConfig = {
            extra-substituters = [
              "https://cache.garnix.io"
            ];
            extra-trusted-public-keys = [
              "cache.garnix.io:CTFPyKSLcx5RMJKfLo5EEPUObbA78b0YQ2DTCJXqr9g="
            ];
          };

          nixosModules = rec {
            default = vpod;
            vpod =
              { pkgs, ... }:
              {
                imports = [ ./nix/service.nix ];
                services.vpod.package = withSystem pkgs.stdenv.hostPlatform.system (
                  { config, ... }: config.packages.vpod
                );
              };
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
            inherit (pkgs) sqlite;
            yt-dlp = inputs.yt-dlp.packages.${system}.default;

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
                shortRev = top.self.shortRev or "dirty";
                version = "${builtins.substring 0 8 lastModifiedDate}-${shortRev}";
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
          };
      }
    );
}

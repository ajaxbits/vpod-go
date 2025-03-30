{
  description = "vpod -- beware the YouTube to podcast feed pipeline";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.latest.url = "github:NixOS/nixpkgs/master";

  outputs =
    {
      nixpkgs,
      latest,
      ...
    }:
    let
      # lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      # version = builtins.substring 0 8 lastModifiedDate;

      # System types to support.
      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
      latestFor = forAllSystems (system: import latest { inherit system; });

    in
    {

      # Add dependencies that are only needed for development
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
          pkgsLatest = latestFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              go-tools
              gopls
              gotools
              just
              pkgsLatest.yt-dlp
              sqlite
              visidata
            ];
          };
        }
      );
    };
}

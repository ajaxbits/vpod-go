{ pkgs, runtimeDeps, ... }:
let
  inherit (pkgs) mkShell;
in
mkShell {
  buildInputs =
    runtimeDeps
    ++ (with pkgs; [
      go
      go-tools
      gopls
      gotools
      just
      sqlc
      sqlite
      vector
      visidata
      yq
    ]);
}

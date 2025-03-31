{
  dockerTools,
  lib,
  package,
  ...
}:
# buildLayeredImage won't work right now due to fakeroot bug
# https://github.com/NixOS/nixpkgs/issues/327311#issuecomment-2525191429
dockerTools.buildImage {
  inherit (package) name;
  tag = package.version;
  copyToRoot = [ dockerTools.caCertificates ];
  config.Cmd = [ (lib.getExe package) ];
}

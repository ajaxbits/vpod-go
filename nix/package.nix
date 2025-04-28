{
  buildGoModule,
  lib,
  makeWrapper,
  name,
  runtimeDeps,
  version,
  ...
}:
let
  fs = lib.fileset;
  gitFiles = fs.gitTracked ../.;
  excludes = fs.unions [
    ../justfile
    ../nix
    ../vector.yaml
  ];
  fileset = fs.difference gitFiles excludes;
  src = fs.toSource {
    inherit fileset;
    root = ../.;
  };
in
buildGoModule {
  inherit src version;
  pname = name;

  vendorHash = "sha256-+ZGm7y7wuukivXBf7cEhJBSJszxTDbqch4Jmyi9mB7M=";

  nativeBuildInputs = [ makeWrapper ];
  preBuild = ''
    go test ./...
  '';
  postFixup = ''
    wrapProgram $out/bin/${name} \
      --set PATH ${lib.makeBinPath runtimeDeps}
  '';

  meta = with lib; {
    description = "Beware the pipeline.";
    homepage = "https://github.com/ajaxbits/vpod-go";
    license = licenses.unlicense;
    mainProgram = name;
    maintainers = with maintainers; [ ajaxbits ];
  };
}

{
  lib,
  name,
  pkgs,
  runtimeDeps,
  version,
  ...
}:
let
  inherit (pkgs) buildGoModule makeWrapper;
  fs = lib.fileset;
  src = fs.toSource {
    root = ../.;
    fileset = fs.gitTracked ../.;
  };
in
buildGoModule {
  inherit src version;
  pname = name;

  vendorHash = "sha256-38AxdH1xuUdLw8TWMlS6N7CIW393W5UyARbCzNVDRDI=";

  nativeBuildInputs = [ makeWrapper ];
  postFixup = ''
    wrapProgram $out/bin/${name} \
      --set PATH ${lib.makeBinPath runtimeDeps}
  '';

  meta = with lib; {
    description = "Beware the pipeline.";
    homepage = "https://github.com/ajaxbits/vpod-go";
    license = licenses.unlicense;
    maintainers = with maintainers; [ ajaxbits ];
  };
}

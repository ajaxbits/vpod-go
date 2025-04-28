{
  config,
  lib,
  vpod,
  ...
}:
let
  inherit (lib) types;
  cfg = config.services.vpod;
in
{
  options.services.vpod = {
    enable = lib.mkEnableOption "vpod";
    hostname = lib.mkOption {
      type = types.string;
      default = "http://localhost";
      description = "Hostname to listen on (with protocol)";
    };
    port = lib.mkOption {
      type = types.port;
      default = 8123;
      description = "Port to listen on";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.vpod = {
      description = "vpod server daemon";
      serviceConfig = {
        ExecStart =
          let
            port = builtins.toString cfg.port;
            baseUrl = cfg.hostname + port;
          in
          ''
            ${lib.getExe vpod} \
              --base-url=${baseUrl} \
              --port=${port}
              --log-level=DEBUG
          '';
      };
      wantedBy = [ "multi-user.target" ];
    };
  };
}

{
  config,
  lib,
  pkgs,
  ...
}:
let
  inherit (lib) types;
  inherit (lib.attrsets) optionalAttrs;
  inherit (lib.lists) elemAt;
  inherit (lib.meta) getExe;
  inherit (lib.modules) mkIf;
  inherit (lib.options) mkOption mkEnableOption;
  inherit (lib.strings) optionalString splitString;

  cfg = config.services.vpod;

  lokicfg = config.services.loki;
  vlcfg = config.services.victorialogs;

  vpod = getExe cfg.package;
  vector = getExe pkgs.vector;
  startScript = pkgs.writeShellScript "vpod-start" (
    vpod
    + optionalString (
      lokicfg.enable && vlcfg.enable
    ) "| ${vector} --config ${cfg.settings.monitoring.vectorConfigPath}"
  );
in
{
  options.services.vpod = {
    enable = mkEnableOption "vpod service";
    package = mkOption {
      type = types.package;
      default = pkgs.vpod or (throw "vpod package not found. Ensure your flake provides it.");
      description = "The vpod package to use.";
    };
    user = mkOption {
      type = types.str;
      default = "vpod";
      description = "User to run the vpod service as.";
    };
    group = mkOption {
      type = types.str;
      default = "vpod";
      description = "Group to run the vpod service as.";
    };

    settings = {
      baseUrl = mkOption {
        type = types.str;
        description = ''
          **Required.** Base URL for the vpod application (e.g., http://localhost:3000).
        '';
      };
      host = mkOption {
        type = types.str;
        default = "0.0.0.0";
        description = "Host address for the vpod application to bind to.";
      };
      port = mkOption {
        type = types.port;
        default = 3000;
        description = "Port for the vpod application to listen on.";
      };

      monitoring = {
        vectorConfigPath = mkOption {
          type = types.path;
          default = "${cfg.package}/etc/vector.yaml";
          description = "Path to the vector configuration file.";
        };
      };

      frontend = {
        user = mkOption {
          type = types.str;
          default = "admin";
          description = "Application-level user for vpod (USER env var).";
        };
        passwordFile = mkOption {
          type = types.path; # TODO: make this the non-store password file
          default = "";
          description = "Path to a file containing the application-level password (PASSWORD_FILE env var).";
        };
      };
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.settings.baseUrl != "";
        message = "services.vpod.baseUrl is required.";
      }
    ];

    users.users.${cfg.user} = {
      inherit (cfg) group;
      isSystemUser = true;
      description = "vpod service user";
    };
    users.groups.${cfg.group} = { };

    systemd.services.vpod = {
      description = "vpod";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        ExecStart = startScript;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = "10s";
        StateDirectory = "vpod";
        Type = "simple";
        User = cfg.user;
        WorkingDirectory = "/var/lib/vpod";
      };

      environment =
        {
          BASE_URL = cfg.settings.baseUrl;
          HOST = cfg.settings.host;
          PASSWORD_FILE = cfg.settings.frontend.passwordFile;
          PORT = builtins.toString cfg.settings.port;
          USER = cfg.settings.frontend.user;
        }
        // optionalAttrs (lokicfg.enable && vlcfg.enable) (
          let
            lokiAddr = lokicfg.configuration.server.http_listen_address or "";
            lokiPort = builtins.toString lokicfg.configuration.server.http_listen_port;
            lokiUrl = if lokiAddr != "" then "${lokiAddr}:${lokiPort}" else "http://0.0.0.0:${lokiPort}";

            vlogsParts = splitString ":" vlcfg.listenAddress;
            vlogsUrl =
              if (elemAt vlogsParts 0 == "") then "http://0.0.0.0${vlcfg.listenAddress}" else vlcfg.listenAddress;
          in
          {
            LOKI = lokiUrl;
            VLOGS = vlogsUrl;
          }
        );
    };
  };
}

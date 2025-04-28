{
  self,
  pkgs,
}:
pkgs.nixosTest {
  name = "vpod service test";

  nodes = {
    # Machine 1: The server that will run the service
    server =
      { config, ... }:
      {
        imports = [ self.nixosModules.vpod ];

        # Configure the service
        services.vpod.enable = true;

        networking.firewall.allowedTCPPorts = [
          config.services.vpod.port
        ];
      };

    # Machine 2: The client that will connect to the vpod service
    # netcat (nc) is already available in the 'libressl' package.
    client = { ... }: { };
  };

  globalTimeout = 20; # Test time limit

  testScript =
    { nodes, ... }:
    ''
      PORT = ${builtins.toString nodes.server.services.vpod.port}
      TEST_STRING = "Test string. The server should echo it back."

      start_all()

      # Wait until VMs are up and the service is started.
      server.wait_for_unit("vpod.service")
      server.wait_for_open_port(${builtins.toString nodes.server.services.vpod.port})
      client.wait_for_unit("network-online.target")

      # The actual test sends an arbitrary string and expects to find it in the output
      output = client.succeed(f"echo '{TEST_STRING}' | nc -4 -N server {PORT}")
      assert TEST_STRING in output
    '';
}

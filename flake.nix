{
  description = "gotelemetry flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in {
      packages = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.buildGoModule {
          pname = "gotelemetry";
          version = "0.81-core";
          
          # Compiles and moves to de nix store. 
          src = ./.;

          vendorHash = "sha256-g+yaVIx4jxpAQ/+WrGKxhVeliYx7nLQe/zsGpxV4Fn4="; 

        env.CGO_ENABLED = "0";

        # Version 
        ldflags = [
            "-s"
            "-w"
            "-X main.BuildVersion=0.81-core"
        ];

          meta = with pkgs.lib; {
            description = "Robust monolithic daemon for telemetry and power consumption monitoring designed for Linux";
            homepage = "https://github.com/JoseM656/gotelemetry";
            license = licenses.bsd3;
            platforms = platforms.unix;
          };
        };
      });

    # For development 
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
          ];
        };
      });
    };
}
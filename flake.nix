{
  description = "gotelemetry flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }@inputs:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      
      # Search a file 'VERSION' in the proyect root
      version = nixpkgs.lib.strings.trim (builtins.readFile ./VERSION);

      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in {
      packages = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.buildGoModule {
          pname = "gotelemetry";
          
          inherit version;
          
          src = ./.;

          vendorHash = "sha256-g+yaVIx4jxpAQ/+WrGKxhVeliYx7nLQe/zsGpxV4Fn4="; 

          env.CGO_ENABLED = "0";

          ldflags = [
            "-s"
            "-w"
            "-X main.BuildVersion=${version}"
          ];

          meta = with pkgs.lib; {
            description = "Robust monolithic daemon for telemetry and power consumption monitoring designed for Linux";
            homepage = "https://github.com/JoseM656/gotelemetry";
            license = licenses.bsd3;
            platforms = platforms.unix;
          };
        };
      });

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
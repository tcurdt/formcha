{
  description = "formcha - ALTCHA server implementation";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "formcha";
            version = "0.2.0";
            src = ./.;

            vendorHash = "sha256-YgtU5OlL2G/dK56eE/jMo6Lx/Y3KS5Xuzjpt5ff8w38=";

            meta = with pkgs.lib; {
              description = "formcha - an ALTCHA server implementation";
              license = licenses.mit;
              maintainers = [ ];
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
          ];
        };
      }
    );
}

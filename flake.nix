{
  description = "Incrementally sync Discord channels to local SQLite";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-darwin" "x86_64-darwin" "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "dissync";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-qOWInVJQ9t9rODdzpKeVeFhJhuR3gEa76TV1g9OD/lg=";
          meta.description = "Incrementally sync Discord channels to local SQLite";
        };
      });
    };
}

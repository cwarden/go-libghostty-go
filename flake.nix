{
  description = "go-libghostty";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";
    flake-utils.url = "github:numtide/flake-utils";
    zig = {
      url = "github:mitchellh/zig-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    nixpkgs,
    flake-utils,
    zig,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShells.default = pkgs.mkShell {
          # Building and testing need only Go. Zig is for `make generate`,
          # which translates libghostty-vt to Go (see generator.go).
          packages = [
            pkgs.go
            pkgs.pinact
            zig.packages.${system}."0.17.0"
          ];
        };
      }
    );
}

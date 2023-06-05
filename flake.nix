{
  description = "Astarte command line client utility";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    flake-utils.url = github:numtide/flake-utils;
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ self.overlays.${system}.default ];
        };
      in
      {
        overlays.default = final: prev: {
          astartectl = final.callPackage ./default.nix { };
        };
        overlay = self.overlays.default;
        packages = { inherit (pkgs) astartectl; };
        packages.default = self.packages.${system}.astartectl;
      }
    );
}

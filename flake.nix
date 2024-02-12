{
  description = "Astarte command line client utility";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    go-utils = { url = "github:noaccOS/go-utils"; inputs.nixpkgs.follows = "nixpkgs"; };
    flake-utils.url = "github:numtide/flake-utils";
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, flake-utils, go-utils, ... }:
    flake-utils.lib.eachSystem go-utils.lib.defaultSystems
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ self.overlays.default ];
          };
        in
        {
          devShells.default = pkgs.simpleGoShell;
        }
      ) // {
      overlays.default = go-utils.lib.asdfOverlay { src = ./.; };
    };
}

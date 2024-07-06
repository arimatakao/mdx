{
  description = "A command-line interface program for downloading manga from the MangaDex website";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
  };
  outputs =
    inputs@{ flake-parts, nixpkgs, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {

      systems = nixpkgs.lib.systems.flakeExposed;

      perSystem =
        { pkgs, ... }:
        let
          mdx = pkgs.callPackage ./nix { };
        in
        {
          formatter = pkgs.nixfmt-rfc-style;

          devShells.default = pkgs.mkShell { inputsFrom = [ mdx ]; };

          packages = {
            default = mdx;
            inherit mdx;
          };
        };
    };
}

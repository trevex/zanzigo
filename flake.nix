{
  description = "zanzigo";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    let
      overlays = [
        # none, yet
      ];
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system overlays;
          config.allowUnfreePredicate = pkg: builtins.elem (nixpkgs.lib.getName pkg) [ "terraform" ];
        };
      in
      rec {
        devShell = pkgs.mkShell rec {
          name = "zanzigo";

          buildInputs = with pkgs; [
            go
            golangci-lint
            delve
            gnumake
            buf
            protobuf
            protoc-gen-go
            protoc-gen-connect-go
            # for bench:
            # terraform
            # google-cloud-sdk
          ];
        };
      }
    );
}

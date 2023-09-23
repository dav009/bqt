{
  description = "bqt (Bigquery test util)";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  inputs.flake-compat = {
    url = "github:edolstra/flake-compat";
    flake = false;
  };

  outputs = { self, nixpkgs, ... }:
    let
      # Generate a user-friendly version number.
      version = builtins.substring 0 8 self.lastModifiedDate;
      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

    in {

      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in rec {
          bqt = pkgs.buildGoModule {
            pname = "bqt";
            inherit version;
            src = ./.;
            CGO_ENABLED = 1;
            preBuild = ''
             export CXX="clang++";
             export CC="clang";
            '';
            vendorSha256 =
              "sha256-sjg+D0IIErl21HZjXBNKBTqXBZfy6w6EhHYS0seUE3k=";
            nativeCheckInputs = with pkgs; [ clang_14 ];
          };
          default = bqt;
        });

      apps = forAllSystems (system: rec {
        bqt = {
          type = "app";
          program = "${self.packages.${system}.bqt}/bin/bqt";
        };
        default = bqt;
      });

      defaultPackage = forAllSystems (system: self.packages.${system}.default);

      defaultApp = forAllSystems (system: self.apps.${system}.default);

      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go_1_21
              gopls
              gotools
              golint
              clang_14
              goreleaser
            ];
            GOROOT = "${pkgs.go_1_21}/share/go";
            CGO_ENABLED = "1";
            CXX = "clang++";
            CC = "clang";
          };
        });

      devShell = forAllSystems (system: self.devShells.${system}.default);
    };
}

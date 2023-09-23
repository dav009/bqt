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

    in
    {

      packages = forAllSystems
        (system:
          let
            pkgs = nixpkgsFor.${system};
          in rec {
            bqt = pkgs.buildGoModule {
              pname = "bqt";
              inherit version;
              src = ./.;
              vendorSha256 =
                "sha256-sjg+D0IIErl21HZjXBNKBTqXBZfy6w6EhHYS0seUE3k=";
            };
            default = bqt;
          });


      apps = forAllSystems (system: rec {
        bqt = {
          type = "app";
          program = "${self.packages.${system}.nix_sample}/bin/bqt";
        };
        default = bqt;
      });

      defaultPackage = forAllSystems (system: self.packages.${system}.default);

      defaultApp = forAllSystems (system: self.apps.${system}.default);

      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go_1_21 gopls gotools golint clang_14 ];
            GOROOT = "${pkgs.go_1_21}/share/go";
          };
        });

      devShell = forAllSystems (system: self.devShells.${system}.default);
    };
}

{ buildGoModule, installShellFiles, stdenv, lib }:
buildGoModule rec {
  pname = "astartectl";
  version = "22.11.03";
  src = ./.;

  nativeBuildInputs = [ installShellFiles ];
  vendorSha256 = "sha256-RVWnkbLOXtNraSoY12KMNwT5H6KdiQoeLfRCLSqVwKQ=";

  postInstall = lib.optionalString (stdenv.hostPlatform == stdenv.buildPlatform) ''
    installShellCompletion --cmd astartectl \
      --bash <($out/bin/astartectl completion bash) \
      --fish <($out/bin/astartectl completion fish) \
      --zsh <($out/bin/astartectl completion zsh)
  '';
}

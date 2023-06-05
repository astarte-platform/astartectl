#!/bin/sh
# Bash completion
mkdir -p /etc/bash_completion.d
astartectl completion bash > /etc/bash_completion.d/astartectl

# ZSH completion
mkdir -p /usr/share/zsh/vendor-completions
astartectl completion zsh > /usr/share/zsh/vendor-completions/_astartectl

# Fish completion
mkdir -p /usr/share/fish/completions
astartectl completion fish > /usr/share/fish/completions/astartectl.fish

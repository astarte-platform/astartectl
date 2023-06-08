#!/bin/sh
# Bash completion
rm /etc/bash_completion.d/astartectl || true

# ZSH completion
rm /usr/share/zsh/vendor-completions/_astartectl || true

# Fish completion
rm /usr/share/fish/completions/astartectl.fish || true

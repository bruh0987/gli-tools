#!/usr/bin/env sh
set -eu

install_dir="${1:-$HOME/.local/bin}"
repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
commit="$(git -C "$repo_root" rev-parse --short HEAD 2>/dev/null || echo unknown)"

mkdir -p "$install_dir"
go build -trimpath -ldflags="-s -w -X main.version=0.1.0 -X main.commit=$commit" -o "$install_dir/gli" "$repo_root/cmd/gli"

case ":$PATH:" in
  *":$install_dir:"*)
    echo "Installed gli to $install_dir/gli"
    ;;
  *)
    echo "Installed gli to $install_dir/gli"
    echo "Add this to your shell profile:"
    echo "export PATH=\"\$PATH:$install_dir\""
    ;;
esac

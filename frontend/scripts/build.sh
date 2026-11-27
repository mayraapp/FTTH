#!/usr/bin/env sh
set -eu

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
source_dir="$root_dir/frontend/src"
published_dir="$root_dir/internal/web/dist"

mkdir -p "$published_dir"
cp -R "$source_dir/." "$published_dir/"

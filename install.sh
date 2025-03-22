#!/bin/bash

set -e

echo "🔧 Building GitSnatch..."

go build -o snatch

echo "📦 Moving to ~/.local/bin..."

mkdir -p ~/.local/bin
mv snatch ~/.local/bin/

if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
  echo '📌 ~/.local/bin is not in your $PATH. Please add this to your shell config:'
  echo 'export PATH="$HOME/.local/bin:$PATH"'
else
  echo "✅ Installed! You can now run 'snatch' from anywhere."
fi

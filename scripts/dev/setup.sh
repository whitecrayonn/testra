#!/usr/bin/env bash
set -e

echo "Setting up Testra local development environment..."

# Install JS dependencies + auto-create Python venv
pnpm install

echo "Setup complete. Run 'pnpm dev' to start the entire stack."

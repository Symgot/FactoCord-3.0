#!/usr/bin/env bash
set -euo pipefail

rm -rf ./release

mkdir ./release
mkdir ./release/FactoCord3

# Erstelle benötigte Unterverzeichnisse
mkdir -p ./release/FactoCord3/temp/temp_settings
mkdir -p ./release/FactoCord3/backups

# Kopiere alle benötigten Dateien
cp config-example.json control.lua FactoCord3 INSTALL.md LICENSE README.md SECURITY.md COMMANDS.md ./release/FactoCord3

# Erstelle leere verification.json falls nicht vorhanden
echo '{"discord_to_factorio":{},"factorio_to_discord":{}}' > ./release/FactoCord3/verification.json

pushd ./release >/dev/null

pushd ./FactoCord3 >/dev/null
chmod 664 ./*
chmod +x ./FactoCord3
chmod 755 ./temp ./temp/temp_settings ./backups
popd >/dev/null

zip -q ./FactoCord3.zip -r ./FactoCord3
echo "Created .zip archive"
tar -czf ./FactoCord3.tar.gz ./FactoCord3
echo "Created .tar archive"
popd >/dev/null

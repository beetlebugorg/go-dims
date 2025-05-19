#!/bin/sh

# This script is used by Dockerfile.build to generate an SBOM for Alpine packages that
# contain binaries that are statically linked in the final build.

echo '{ "bomFormat": "CycloneDX", "specVersion": "1.4", "version": 1, "components": ['

FIRST=1
for pkg in "$@"; do
  info=$(apk info "$pkg" -d -P --license -w)
  version=$(echo "$info" | head -1 | sed -E "s/^$pkg-(.*) description:/\1/")
  license=$(echo "$info" | grep -A1 license: | tail -1)
  webpage=$(echo "$info" | grep -A1 webpage: | tail -1)

  [ "$FIRST" = 0 ] && echo ',' || FIRST=0

  cat <<EOF
{
  "type": "library",
  "name": "$pkg",
  "version": "$version",
  "licenses": [ { "license": { "id": "$license" } } ],
  "externalReferences": [ { "type": "website", "url": "$webpage" } ],
  "purl": "pkg:apk/alpine/$pkg@$version"
}
EOF
done

echo ']}'
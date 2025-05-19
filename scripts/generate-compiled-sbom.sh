#!/bin/sh

# This script is used by Dockerfile.build to generate an SBOM for source packages that
# are compiled and statically linked in the final build.

set -e

while [ $# -gt 0 ]; do
  case "$1" in
    --name) NAME="$2"; shift 2 ;;
    --version) VERSION="$2"; shift 2 ;;
    --license) LICENSE="$2"; shift 2 ;;
    --website) WEBSITE="$2"; shift 2 ;;
    --download) DOWNLOAD="$2"; shift 2 ;;
    --checksum) CHECKSUM="$2"; shift 2 ;;
    --license_file) LICENSE_FILE="$2"; shift 2 ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
done

if [ -z "$NAME" ] || [ -z "$VERSION" ] || [ -z "$LICENSE" ] || [ -z "$WEBSITE" ]; [ -z "$DOWNLOAD" ]; [ -z "$CHECKSUM" ]; [ -z "$LICENSE_FILE" ];  then
  echo "Usage: $0 --name NAME --version VERSION --license LICENSE --website WEBSITE --download DOWNLOAD --checksum CHECKSUM --license_file LICENSE_FILE"
  exit 1
fi

CREATED=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
DOC_NS="https://${WEBSITE}/spdx/${NAME}-${VERSION}"

cat <<EOF
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "version": 1,
  "components": [
    {
      "type": "library",
      "name": "${NAME}",
      "version": "${VERSION}",
      "licenses": [ { "license": { "id": "${LICENSE}" } } ],
      "externalReferences": [ { "type": "website", "url": "${WEBSITE}" } ],
      "purl": "pkg:sourcearchive/${NAME}@${VERSION}?download_url=${DOWNLOAD}&checksum=sha256:${CHECKSUM}"
    }
  ],
  "externalReferences": [
    {
      "type": "source-distribution",
      "url": "$DOWNLOAD",
      "comment": "Upstream source repository",
      "license": "${LICENSE_FILE}",
      "hashes": [
        {
          "algorithms": ["sha256"],
          "hash": "${CHECKSUM}"
        }
      ]
    }
  ]
}
EOF
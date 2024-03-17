#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Source
ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/../..
source "${ROOT_PATH}/scripts/install/docker.sh"

# Params
tag=$1

# env
export APP_VERSION=$tag

# Validate params
validate_params() {
  if [ -z "$tag" ]; then
    echo "validation failed: tag is empty."
    exit 1
  fi
}

# Build
build() {
  echo "start building..."

  cd build/docker || exit
  tar -czvpf docker-compose.tar.gz * .env.example

  cp .env.example .env
  docker-compose build
  docker save bingo-apiserver:"$tag" | gzip >bingo-apiserver.tar.gz

  cd - || exit
  mkdir -p _output
  mv build/docker/*.tar.gz _output/

  ls -lh

  echo "build success"
}

# Run
validate_params "$1"
build

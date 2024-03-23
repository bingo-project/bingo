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

  # list images to save.
  images=("bingo-apiserver" "bingo-watcher" "bingo-bot" "bingoctl")

  # Add tag
  for index in "${!images[@]}"
  do
    images[index]="${images[index]}:${tag}"
  done

  cd build/docker || exit
  tar -czvpf bingo-docker.tar.gz * .env.example

  cp .env.example .env
  docker-compose build
  docker save "${images[@]}" | gzip >bingo-images.tar.gz

  cd - || exit
  mkdir -p _output
  mv build/docker/*.tar.gz _output/

  ls -lh

  echo "build success"
}

# Run
validate_params "$1"
build

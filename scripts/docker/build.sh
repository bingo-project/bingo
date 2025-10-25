#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Source
ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/../..

# Copy env
if [ ! -f "${ROOT_PATH}/scripts/docker/env.sh" ]; then
    # 如果不存在，复制 env.example.sh 为 env.sh
    cp "${ROOT_PATH}/scripts/docker/env.example.sh" "${ROOT_PATH}/scripts/docker/env.sh"
    echo "env.sh not found, create from env.example.sh"
else
    echo "env.sh found"
fi

source "${ROOT_PATH}/scripts/install/init.sh"
source "${ROOT_PATH}/scripts/docker/env.sh"

# Usage
usage() {
  echo "Usage: $0 [-n <name>] [-i <images>] [-a <architecture>] [-h]"
  echo "  -n <name> App name, default: current directory name"
  echo "  -i <images> Images to build, default: all"
  echo "  -a <architecture> linux architecture, default: amd64, support: amd64, arm64"
  echo "  -f <build_from> build from, default: bin, support: image, bin"
  echo "  -h Show help"
  exit 1
}

# Check params
if [ $# -eq 0 ]; then
  usage
fi

# Parse params
while getopts "n:i:a:h" opt; do
  case $opt in
  n)
    app_name=$OPTARG
    ;;
  i)
    # Read images
    IFS=','
    images=$OPTARG
    ;;
  a)
    architecture=$OPTARG
    ;;
  f)
    build_from=$OPTARG
    ;;
  h)
    usage
    ;;
  ?)
    usage
    ;;
  esac
done

# Build
build() {
  export APP_VERSION=$version
  export IMAGE_PLATFORM=linux/$architecture

  # Add tag to image
  for index in "${!images[@]}"; do
    images[index]="${registry_prefix}/${images[index]}:${APP_VERSION}"
  done

  # App info
  echo "app: $app_name"
  echo "image: ${images[@]}"
  echo "architecture: $architecture"
  echo "start building..."

  # images
  cd deployments/docker || exit
  cp .env.example .env

  if [ "$build_from" = "bin" ]; then
      file="docker-compose.bin.yaml"
  else
      file="docker-compose.yaml"
  fi

  docker-compose -f $file build

  echo "build success"
}

# Run
build

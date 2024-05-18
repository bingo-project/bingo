#!/usr/bin/env bash

# Source
ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/../..
source "${ROOT_PATH}/scripts/docker/env.sh"

# Params
remote_user=$1
remote_host=$2
tag=$3

# Validate params
validate_params() {
  # Default params
  if [ -z "$1" ]; then
    remote_user=$SSH_USER
  fi
  if [ -z "$2" ]; then
    remote_host=$SSH_HOST
  fi
  if [ -z "$3" ]; then
    tag="$version"
  fi

  if [ -z "$remote_user" ]; then
    echo "validation failed: remote_user is empty."
    exit 1
  fi
  if [ -z "$remote_host" ]; then
    echo "validation failed: remote_host is empty."
    exit 1
  fi
}

package() {
  cd build/docker || exit

  # config file
  mkdir -p config _output
  cp -a ../../configs/*.yaml config

  # docker-compose
  tar -czvpf "$app_name"-docker.tar.gz * .env.example

  rm -r config
  cd - || exit
  mv build/docker/*.tar.gz _output/

  ls -lh _output/
}

scp_to_remote() {
  echo "start scp to remote..."
  scp -o StrictHostKeyChecking=no _output/*.tar.gz "$remote_user"@"$remote_host":/tmp/
  # rm _output/*.tar.gz
}

deploy_to_remote() {
  echo "start deploy to remote..."
  ssh -o StrictHostKeyChecking=no "$remote_user"@"$remote_host" 'bash -s' "$app_name" "$tag" <./scripts/docker/run.sh
}

# Run
validate_params "$1" "$2" "$3"
package
scp_to_remote
deploy_to_remote

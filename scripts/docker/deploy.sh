#!/usr/bin/env bash

# Source
ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/../..
source "${ROOT_PATH}/scripts/lib/ssh-agent.sh"
source "${ROOT_PATH}/scripts/lib/env.sh"

# Params
tag=$1
remote_user=$2
remote_host=$3
remote_key=$4

# Validate params
validate_params() {
  # Default params
  if [ -z "$1" ]; then
    tag="$version"
  fi
  if [ -z "$2" ]; then
    remote_user=$SSH_USER
  fi
  if [ -z "$3" ]; then
    remote_host=$SSH_HOST
  fi
  if [ -z "$4" ]; then
    remote_key=$SSH_KEY
  fi

  if [ -z "$remote_user" ]; then
    echo "validation failed: remote_user is empty."
    exit 1
  fi
  if [ -z "$remote_host" ]; then
    echo "validation failed: remote_host is empty."
    exit 1
  fi

  #  if [ -z "$remote_key" ]; then
  #    echo "validation failed: remote_key is empty."
  #    exit 1
  #  fi
}

package() {
  cd build/docker || exit

  # config file
  mkdir config
  cp -a ../../configs/*.example.yaml config

  # docker-compose
  tar -czvpf "$app_name"-docker.tar.gz * .env.example

  rm -r config
  cd - || exit
  mv build/docker/*.tar.gz _output
}

scp_to_remote() {
  scp _output/*.tar.gz "$remote_user"@"$remote_host":/tmp/
  # rm _output/*.tar.gz
}

deploy_to_remote() {
  ssh "$remote_user"@"$remote_host" 'bash -s' "$app_name" "$tag" <./scripts/docker/run.sh
}

# Run
validate_params "$1" "$2" "$3" "$4"
start_ssh_agent
package
scp_to_remote
deploy_to_remote

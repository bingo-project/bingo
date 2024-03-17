#!/usr/bin/env bash

# Params
remote_host=$1
remote_key=$2

start_ssh_agent() {
  which ssh-agent || (apt-get update -qq && apt-get -y -qq install openssh-client)
  eval "$(ssh-agent -s)"

  mkdir -p ~/.ssh
  touch ~/.ssh/config
  touch ~/.ssh/known_hosts
  chmod -R 400 ~/.ssh

  cat "$remote_key" | ssh-add -
  ssh-add -l

  ssh-keyscan "$remote_host" >> ~/.ssh/known_hosts
  echo "StrictHostKeyChecking no" >>~/.ssh/config
}

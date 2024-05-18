#!/bin/bash

# Git
install_git() {
  if ! command -v git &>/dev/null; then
    apt update && apt install -y -qq git
  fi
}

# Docker
install_docker() {
  if ! command -v docker &>/dev/null; then
    apt update && apt install -y -qq docker.io
  fi

  if ! command -v docker-compose &>/dev/null; then
    curl -SL https://github.com/docker/compose/releases/download/v2.27.0/docker-compose-linux-x86_64 \
              -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose && ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
  fi
}

# Start install
install_git
install_docker

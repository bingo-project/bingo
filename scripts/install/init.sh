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
    apt update && apt install -y -qq docker.io docker-compose
  fi
}

# Start install
install_git
install_docker

#!/bin/bash

# Docker
install_docker() {
  which docker || apt-get install -y -qq docker.io docker-compose
}

# Start install
install_docker

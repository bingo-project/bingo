#!/usr/bin/env bash

# Params
tag=$1

run() {
  mkdir -p /opt/bingo
  cd /opt/bingo || exit

  tar -xzvpf /tmp/docker-compose.tar.gz -C ./
  if [ ! -f .env ]; then
    cp .env.example .env
  fi
  if [ ! -f config/bingo-apiserver.yaml ]; then
    cp config/bingo-apiserver.example.yaml config/bingo-apiserver.yaml
  fi

  # Update app version by .env
  if [ -n "${tag}" ]; then
    sed -i "s/APP_VERSION=.*/APP_VERSION=${tag}/g" .env
  fi

  docker load </tmp/bingo-apiserver.tar.gz
  docker-compose up -d

  rm /tmp/docker-compose.tar.gz
  rm /tmp/bingo-apiserver.tar.gz
  rm config/bingo-apiserver.example.yaml
}

run

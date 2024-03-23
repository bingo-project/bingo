#!/usr/bin/env bash

# Params
tag=$1

run() {
  app_name=bingo
  mkdir -p /opt/${app_name}
  cd /opt/${app_name} || exit

  tar -xzvpf /tmp/${app_name}-docker.tar.gz -C ./
  if [ ! -f .env ]; then
    cp .env.example .env
  fi
  if [ ! -f config/${app_name}-apiserver.yaml ]; then
    cp config/${app_name}-apiserver.example.yaml config/${app_name}-apiserver.yaml
  fi
  if [ ! -f config/promtail.yaml ]; then
    cp config/promtail.example.yaml config/promtail.yaml
  fi

  # Update app version by .env
  if [ -n "${tag}" ]; then
    sed -i "s/APP_VERSION=.*/APP_VERSION=${tag}/g" .env
  fi

  # Load and tag latest
  loaded=$(docker load </tmp/${app_name}-images.tar.gz)
  for image_with_version in $(echo "$loaded" | awk -F ': ' '{print $2}'); do
    image=${image_with_version%:*}
    docker tag "$image_with_version" "$image":latest
  done

  docker-compose up -d

  rm /tmp/${app_name}*.tar.gz
  rm config/${app_name}-apiserver.example.yaml
  rm config/promtail.example.yaml
}

run

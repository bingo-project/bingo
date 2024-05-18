# ==============================================================================
# Makefile helper functions for docker image
#

DOCKER := docker

# Docker version
TAG := $(VERSION:v%=%)

# Image build related variables.
REGISTRY_PREFIX ?= bingo

# Determine image files by looking into cmd/*
IMAGES_DIR ?= $(wildcard ${ROOT_DIR}/cmd/*)
# Determine images names by stripping out the dir names
IMAGES ?= $(filter-out tools,$(foreach image,${IMAGES_DIR},$(notdir ${image})))

ifeq (${IMAGES},)
  $(error Could not determine IMAGES, set ONEX_ROOT or run in source dir)
endif

.PHONY: image.build
image.build:
	$(ROOT_DIR)/scripts/docker/build.sh -a amd64

.PHONY: image.push
image.push: $(addprefix image.push., $(IMAGES))

.PHONY: image.push.%
image.push.%: image.build.% ## Build and push specified docker image.
	$(eval IMAGE := $*)
	@echo "===========> Pushing image $(IMAGE) $(TAG) to $(REGISTRY_PREFIX)"
	$(DOCKER) push $(REGISTRY_PREFIX)/$(IMAGE):$(TAG)

SHELL := /bin/bash # Use bash syntax

# Set up variables
GO111MODULE=on

# Build ldflags
VERSION ?= v0.0.0
GITCOMMIT=$(shell git rev-parse HEAD)
BUILDDATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS=-ldflags "-X main.version=$(VERSION) \
			-X main.buildHash=$(GITCOMMIT) \
			-X main.buildDate=$(BUILDDATE)"

AWS_SERVICE=$(shell echo $(SERVICE) | tr '[:upper:]' '[:lower:]')
CONTAINER_REPOSITORY ?= public.ecr.aws/aws-controllers-k8s/controller

.PHONY: all test

all: test

test: 				## Run code tests
	go test -v ./...

build-controller-image: ## Build controller container image
	docker build \
		--build-arg service_alias=${AWS_SERVICE} \
		--tag $(CONTAINER_REPOSITORY):$(AWS_SERVICE)-$(VERSION) .

help:           	## Show this help.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep | sed -e 's/\\$$//' \
		| awk -F'[:#]' '{print $$1 = sprintf("%-30s", $$1), $$4}'

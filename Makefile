SHELL := /bin/bash # Use bash syntax

# Set up variables
GO111MODULE=on

# Build ldflags
VERSION ?= "v0.0.0"
GITCOMMIT=$(shell git rev-parse HEAD)
BUILDDATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS=-ldflags "-X main.version=$(VERSION) \
			-X main.buildHash=$(GITCOMMIT) \
			-X main.buildDate=$(BUILDDATE)"

AUTHENTICATED_ACCOUNT_ID=$(shell aws sts get-caller-identity --output text --query "Account")

.PHONY: all test

all: test

local-run-controller: ## Run a controller image locally for SERVICE
	@go run ./cmd/controller/main.go \
		--aws-account-id=$(AUTHENTICATED_ACCOUNT_ID) \
		--aws-region=us-west-2 \
		--enable-development-logging \
		--log-level=debug

test: 				## Run code tests
	go test -v ./...

help:           	## Show this help.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep | sed -e 's/\\$$//' \
		| awk -F'[:#]' '{print $$1 = sprintf("%-30s", $$1), $$4}'
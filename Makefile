SHELL := /bin/zsh

define PROJECT_HELP_MSG
Usage:
  make help:\t show this message
  make lint:\t run go linter
  make build:\t build c7n-helper binary
endef
export PROJECT_HELP_MSG

help:
	echo -e $$PROJECT_HELP_MSG

lint:
	golangci-lint run

build:
	go build .

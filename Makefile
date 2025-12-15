SHELL := /usr/bin/bash
.ONESHELL:
.SHELLFLAGS := -euo pipefail -c

APP_NAME     ?= your-api
ENV_FILE     ?= .env
COMPOSE      ?= docker compose
COMPOSE_FILE ?= deploy/docker/docker-compose.dev.yml

GO     ?= go
GOFMT  ?= gofmt

include make/audit.mk
include make/boundary.mk
include make/db.mk
include make/dev.mk
include make/go.mk
include make/help.mk
include make/prereq.mk
include make/sanity.mk
include make/sec.mk
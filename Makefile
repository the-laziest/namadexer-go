#!/bin/sh

NAMADA_VERSION := 0.31.1
BASE_URL := https://raw.githubusercontent.com/anoma/namada
URL := $(BASE_URL)/v$(NAMADA_VERSION)/wasm/checksums.json

CHECK_CURL := $(shell command -v curl 2> /dev/null)
CHECK_WGET := $(shell command -v wget 2> /dev/null)

CANT_DOWNLOAD=0

ifdef CHECK_CURL
DOWNLOAD_CMD = curl -L -o
else ifdef CHECK_WGET
DOWNLOAD_CMD = wget -O
else
CANT_DOWNLOAD=1
endif

download-checksum:
	if [ ! -f checksums.json ]; then \
		if [ $(CANT_DOWNLOAD) -eq 1 ]; then \
			echo "Neither curl nor wget are available on your system"; \
			exit 1; \
		fi; \
		echo $(URL); \
		$(DOWNLOAD_CMD) checksums.json $(URL); \
	fi

run-postgres:
	docker run --name postgres -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=blockchain -p 5432:5432 -d postgres:14

run-indexer: download-checksum
	CONFIG_PATH="${PWD}/config/config.toml" go run ./cmd/indexer/

run-server: download-checksum
	CONFIG_PATH="${PWD}/config/config.toml" go run ./cmd/server/

compose: download-checksum
	docker compose -f docker-compose.yaml up -d postgres
	sleep 5
	docker compose -f docker-compose.yaml up -d --build indexer server

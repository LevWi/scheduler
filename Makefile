# ==============================
# Project configuration
# ==============================

PROJECT_ROOT := $(shell pwd)
MODULE_DIR := back/appointment-service
BOT_DIR := $(MODULE_DIR)/cmd/bot
SERVICE_DIR := $(MODULE_DIR)/cmd/service

BINARY_NAME := bot
BUILD_DIR := bin

# ==============================
# Default target
# ==============================

.PHONY: help build test run_bot run_service
help:
	@echo "Available commands:"
	@echo "  make run_*   - Run bot/service/all locally"
	@echo "  make build   - Build binary"
	@echo "  make test    - Run tests"
	@echo "  make fmt     - Format code"
	@echo "  make vet     - Run go vet"
	@echo "  make clean   - Remove build artifacts"

# ==============================
# Run
# ==============================

CONFIG_PATH := $(PROJECT_ROOT)

run_bot:
	SCHED_CONFIG_PATH="$(CONFIG_PATH)/bot_cfg.yaml" SHED_LOG_LEVEL=debug go -C $(BOT_DIR) run .

run_service:
	SCHED_CONFIG_PATH="$(CONFIG_PATH)/test_cfg.yaml" SHED_LOG_LEVEL=debug go -C $(SERVICE_DIR) run .

run_all: run_service run_bot

# ==============================
# Build
# ==============================

build:
	mkdir -p $(BUILD_DIR)
	go -C $(BOT_DIR) build -o $(PROJECT_ROOT)/$(BUILD_DIR)/$(BINARY_NAME)

# ==============================
# Test
# ==============================

test:
	go -C $(MODULE_DIR) test ./...

# ==============================
# Code quality
# ==============================

fmt:
	go -C $(MODULE_DIR) fmt ./...

vet:
	go -C $(MODULE_DIR) vet ./...

# ==============================
# Cleanup
# ==============================

# .PHONY: clean
# clean:
# 	rm -rf ./$(BUILD_DIR)/
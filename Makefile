# Makefile for building the dnpm application

# Define variables
GO := go
BUILD_DIR := ./dist

# Targets and their respective commands
.PHONY: build

build:
    @echo "Building dnpm..."
    $(GO) build -o $(BUILD_DIR)/dnpm cmd/dnpm.go
    @echo "Build complete. Binary is located in $(BUILD_DIR)"

# Clean the build directory
clean:
    @echo "Cleaning up..."
    rm -rf $(BUILD_DIR)
    @echo "Cleanup complete."

# Default target, which is executed when you run 'make' without arguments
default: build


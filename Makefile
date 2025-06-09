include .env
export

.PHONY: run
run:
	go run cmd/api/main.go

ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
# ถ้า BUILD_VERSION ไม่ถูกเซ็ตใน .env, ให้ใช้ git tag ล่าสุด (ถ้าไม่มี tag จะ fallback เป็น "unknown")
BUILD_VERSION := $(or ${BUILD_VERSION}, $(shell git describe --tags --abbrev=0 2>/dev/null || echo "unknown"))
BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S%z")

.PHONY: build
build:
	go build -ldflags \
	"-X 'go-mma/build.Version=${BUILD_VERSION}' \
	-X 'go-mma/build.Time=${BUILD_TIME}'" \
	-o app cmd/api/main.go

.PHONY: image
image:
	docker build \
	-t go-mma:${BUILD_VERSION} \
	--build-arg VERSION=${BUILD_VERSION} \
	.
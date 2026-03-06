LIB_NAME = libgogit
SRC = ./shared

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
DEV_VERSION := 0.1.1-dev.$(shell date +%Y%m%d%H%M%S)

ifeq ($(UNAME_S),Darwin)
  ifeq ($(UNAME_M),arm64)
    CURRENT_BUILD := build-darwin-arm64
  else
    CURRENT_BUILD := build-darwin-amd64
  endif
else ifeq ($(UNAME_S),Linux)
  CURRENT_BUILD := build-linux-amd64
else
  CURRENT_BUILD := build-windows-amd64
endif

.PHONY: generate build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-windows-amd64 all pack clean dev

generate:
	cd generate && go run .

build-darwin-arm64:
	@mkdir -p runtimes/osx-arm64/native
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -buildmode=c-shared -o runtimes/osx-arm64/native/$(LIB_NAME).dylib $(SRC)

build-darwin-amd64:
	@mkdir -p runtimes/osx-x64/native
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -buildmode=c-shared -o runtimes/osx-x64/native/$(LIB_NAME).dylib $(SRC)

build-linux-amd64:
	@mkdir -p runtimes/linux-x64/native
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc \
		go build -buildmode=c-shared -o runtimes/linux-x64/native/$(LIB_NAME).so $(SRC)

build-windows-amd64:
	@mkdir -p runtimes/win-x64/native
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
		go build -buildmode=c-shared -o runtimes/win-x64/native/$(LIB_NAME).dll $(SRC)

all: build-darwin-arm64 build-darwin-amd64

pack:
	cd dotnet && dotnet pack -c Release

dev: generate $(CURRENT_BUILD)
	cd dotnet && dotnet pack -c Release /p:Version=$(DEV_VERSION)
	@echo "Packed: dotnet/GoGit.Interop/bin/Release/GoGit.Interop.$(DEV_VERSION).nupkg"

clean:
	rm -rf runtimes/
	rm -f shared/libgogit.h

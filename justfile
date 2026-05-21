set fallback

default:
    just --choose

list:
    just --list


import? 'justfile.custom'

# Variables with defaults
CGO_ENABLED := env_var_or_default('CGO_ENABLED', '0')
GO111MODULE := env_var_or_default('GO111MODULE', 'on')
GOFLAGS := env_var_or_default('GOFLAGS', '"-mod=mod"')
SUFFIX := env_var_or_default('SUFFIX', '')

# Determine project name
NAME := if env_var_or_default('GITEA_REPO_NAME', '') != '' {
    env_var('GITEA_REPO_NAME')
} else {
    file_name(justfile_directory())
}

OUTBIN := env_var_or_default('OUTBIN', NAME)

# Build metadata
VERSION := env_var_or_default('VERSION', `git describe --tags --always --dirty 2> /dev/null || echo dev`)
BUILD_TIME := env_var_or_default('BUILD_TIME', `/bin/date +%FT%T%z`)
PKG := `GOWORK=off go list -m`
LD_FLAGS_OPTIMIZE := env_var_or_default('LD_FLAGS_OPTIMIZE', '-s -w')
LD_FLAGS := '"-X ' + PKG + '/pkg/version.Version=' + VERSION + ' -X ' + PKG + '/pkg/version.BuildTime=' + BUILD_TIME + ' ' + LD_FLAGS_OPTIMIZE + '"'
GO_OPTS := env_var_or_default('GO_OPTS', '-trimpath')
OPTIMIZE := env_var_or_default('OPTIMIZE', 'false')

# Display version information
version:
    @echo {{OUTBIN}} @ {{NAME}} {{VERSION}} {{BUILD_TIME}} {{PKG}} {{LD_FLAGS}}

# Build for all platforms
all-build: build-linux-amd64 build-linux-arm64 build-linux-386

# Download dependencies
download:
    @echo Download go.mod dependencies
    go mod download

# Generate code
generate:
    GOWORK=off go generate -mod=mod ./...

# Run wire dependency injection
wire:
    go generate -x ./di

# Build for specific platform
build GOOS GOARCH:
    GOOS={{GOOS}} GOARCH={{GOARCH}} CGO_ENABLED={{CGO_ENABLED}} go build \
        -ldflags {{LD_FLAGS}} {{GO_OPTS}} \
        -mod readonly \
        -o bin/{{OUTBIN}}{{SUFFIX}}_{{GOOS}}_{{GOARCH}} \
        ./cmd/{{OUTBIN}}/main.go

build-race:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -race \
        -ldflags {{LD_FLAGS}} {{GO_OPTS}} \
        -mod readonly \
        -o bin/{{OUTBIN}}{{SUFFIX}}_linux_amd64 \
        ./cmd/{{OUTBIN}}/main.go

# Install binaries
install:
    CGO_ENABLED={{CGO_ENABLED}} go install \
        -ldflags {{LD_FLAGS}} \
        ./...

# Build for Linux ARM64
build-linux-arm64: (build "linux" "arm64")

# Build for Linux AMD64
build-linux-amd64: (build "linux" "amd64")

# Build for Linux 386
build-linux-386: (build "linux" "386")

# Pre-push checks
prepush: generate vet fmt tidy test

# Run tests
test:
    CGO_ENABLED=0 go build ./...
    CGO_ENABLED=1 go test -race -cover -v -mod=readonly ./... && echo -e "\033[32mSUCCESS\033[0m" || (echo -e "\033[31mFAILED\033[0m" && exit 1)

# Run benchmarks
bench:
    go test -test.timeout=30m -benchmem -run ^$$ -benchtime=20s -bench . ./... && echo -e "\033[32mSUCCESS\033[0m" || (echo -e "\033[31mFAILED\033[0m" && exit 1)

# Run go vet
vet:
    go vet ./...

# Format code
fmt:
    go fmt ./...

# Tidy dependencies
tidy:
    go mod tidy

verify:
    go mod verify

# List all modules
go_list:
    go list -u -m all

# Update all dependencies
go_update_all:
    go get -t -u ./...


BUILD_FILES = $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}}\
{{end}}' ./...)

XT_VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
# XT_VERSION = v1.0.0
LDFLAGS := -X github.com/adamkobi/xt/internal/build.Version=$(XT_VERSION) $(LDFLAGS)

bin/xt: $(BUILD_FILES)
	@go build -trimpath -ldflags "$(LDFLAGS)" -o "$@" ./cmd/main.go

test:
	go test ./...
.PHONY: test
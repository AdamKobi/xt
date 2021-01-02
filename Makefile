BUILD_FILES = $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}}\
{{end}}' ./...)

# XT_VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
XT_VERSION = v1.0.0
LDFLAGS := -X github.com/adamkobi/xt/internal/build.Version=$(XT_VERSION) $(LDFLAGS)
ifdef XT_OAUTH_CLIENT_SECRET
	LDFLAGS := -X github.com/adamkobi/xt/context.oauthClientID=$(XT_OAUTH_CLIENT_ID) $(LDFLAGS)
	LDFLAGS := -X github.com/adamkobi/xt/context.oauthClientSecret=$(XT_OAUTH_CLIENT_SECRET) $(LDFLAGS)
endif

bin/xt: $(BUILD_FILES)
	@go build -trimpath -ldflags "$(LDFLAGS)" -o "$@" ./cmd/main.go

test:
	go test ./...
.PHONY: test
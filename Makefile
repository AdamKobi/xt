BUILD_FILES = $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}}\
{{end}}' ./...)

# XT_VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
XT_VERSION = v1.0.0
LDFLAGS := -X github.com/adamkobi/xt/command.Version=$(XT_VERSION) $(LDFLAGS)
ifdef XT_OAUTH_CLIENT_SECRET
	LDFLAGS := -X github.com/adamkobi/xt/context.oauthClientID=$(XT_OAUTH_CLIENT_ID) $(LDFLAGS)
	LDFLAGS := -X github.com/adamkobi/xt/context.oauthClientSecret=$(XT_OAUTH_CLIENT_SECRET) $(LDFLAGS)
endif

bin/xt: $(BUILD_FILES)
	@go build -trimpath -ldflags "$(LDFLAGS)" -o "$@" ./cmd/main.go

test:
	go test ./...
.PHONY: test

site:
	git clone https://github.com/github/adamkobi.github.com.git "$@"

site-docs: site
	git -C site pull
	git -C site rm 'manual/xt*.md' 2>/dev/null || true
	go run ./cmd/gen-docs site/manual
	for f in site/manual/xt*.md; do sed -i.bak -e '/^### SEE ALSO/,$$d' "$$f"; done
	rm -f site/manual/*.bak
	git -C site add 'manual/xt*.md'
	git -C site commit -m 'update help docs'
.PHONY: site-docs
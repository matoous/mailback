ifndef GOPATH
GOPATH := $(shell go env GOPATH)
export GOPATH
endif

GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)

# Go files to check with linter
GO_FILES := $(shell find . -type f -name '*.go')

# Console size
COLUMNS ?= $(lastword $(shell stty size 2>/dev/null || echo 0 80))

# Colorful output
color_off = \033[0m
color_green = \033[0;32m
color_cyan = \033[0;36m
color_yellow = \033[0;33m

define echo_green
    $(call echo_custom_color,$(color_green),$(1))
endef

define echo_cyan
	$(call echo_custom_color,$(color_cyan),$(1))
endef

define echo_warning
	$(call echo_custom_color,$(color_yellow),$(1))
endef

define echo_custom_color
	@printf "$(1)$(2)$(color_off)\n"
endef

# Before and after job output functions
define before_job
	$(eval TO_WRITE = $(strip $(1)))
	$(eval WRITE_LENGTH = $(shell expr \( 11 + \( "X$(TO_WRITE)" : ".*" \) \) || echo 0 ))
	$(eval TO_WRITE = ╔$(shell f=0; while [ $$((f+=1)) -le 8 ]; do printf ═; done;) $(TO_WRITE))
	$(eval TO_WRITE = $(TO_WRITE) $(shell f=$(WRITE_LENGTH); while [ $$((f+=1)) -le $(COLUMNS) ]; do printf ═; done;)╗)
	$(call echo_cyan,$(TO_WRITE))
	$(eval TO_WRITE = "")
	$(eval WRITE_LENGTH = 0)
endef

define after_job
	$(eval TO_WRITE = $(strip $(1)))
	$(eval WRITE_LENGTH = $(shell expr \( 11 + \( "X$(TO_WRITE)" : ".*" \) \) || echo 0 ))
	$(eval TO_WRITE = ╚$(shell f=0; while [ $$((f+=1)) -le 8 ]; do printf ═; done;) $(TO_WRITE))
	$(eval TO_WRITE = $(TO_WRITE) $(shell f=$(WRITE_LENGTH); while [ $$((f+=1)) -le $(COLUMNS) ]; do printf ═; done;)╝)
	$(call echo_green,$(TO_WRITE))
	$(eval TO_WRITE = "")
	$(eval WRITE_LENGTH = 0)
endef

define warning_job
	$(eval TO_WRITE = $(strip $(1)))
	$(eval WRITE_LENGTH = $(shell expr \( 11 + \( "X$(TO_WRITE)" : ".*" \) \) || echo 0 ))
	$(eval TO_WRITE = ╚$(shell f=0; while [ $$((f+=1)) -le 8 ]; do printf ═; done;) $(TO_WRITE))
	$(eval TO_WRITE = $(TO_WRITE) $(shell f=$(WRITE_LENGTH); while [ $$((f+=1)) -le $(COLUMNS) ]; do printf ═; done;)╝)
	$(call echo_warning,$(TO_WRITE))
	$(eval TO_WRITE = "")
	$(eval WRITE_LENGTH = 0)
endef

# Make this makefile self-documented with target `help`
.PHONY: help
.DEFAULT_GOAL := help

help:
	@grep -Eh '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

configure: ## Prepare environment for compiling: intall dep and download dependencies
	$(call before_job,Installing dependencies:)
	go mod download
	$(call after_job,Dependencies were successfully installed.)

configure-check: configure ## Install linters if they are not presented in $PATH
ifndef REVIVE
	GO111MODULE=off go get -u github.com/mgechev/revive
endif

build-receiver: configure ## Build receiver binary
	$(call before_job,Building binary:)
	go build -o mailback-receiver -a -tags "osusergo netgo" --ldflags "-linkmode external -extldflags '-static'" cmd/receiver/main.go
	$(call after_job,Binary was successfully built.)

build-sender: configure ## Build sender binary
	$(call before_job,Building binary:)
	go build -o mailback-sender -a -tags "osusergo netgo" --ldflags "-linkmode external -extldflags '-static'" cmd/sender/main.go
	$(call after_job,Binary was successfully built.)

build-server: configure ## Build web server binary
	$(call before_job,Building binary:)
	go build -o mailback-server -a -tags "osusergo netgo" --ldflags "-linkmode external -extldflags '-static'" cmd/server/main.go
	$(call after_job,Binary was successfully built.)

test: configure ## Run tests
	$(call before_job,Tests are running:)
	go test -race ./... -cover
	$(call after_job,Tests succeeded!)

lint: configure-check go-mod-tidy revive lint-golangci ## Run all linters

go-mod-tidy: ## Check if go.mod and go.sum does not contains any unnecessary dependencies and remove them.
	$(call before_job,Go mod tidy checking dependencies:)
ifndef TMPDIR
	$(eval TMPDIR=$(shell mktemp -d))
endif
	cp -fv go.mod $(TMPDIR)
	cp -fv go.sum $(TMPDIR)
	go mod tidy -v
	diff -u $(TMPDIR)/go.mod go.mod
	diff -u $(TMPDIR)/go.sum go.sum
	rm -f $(TMPDIR)go.mod $(TMPDIR)go.sum
	$(call after_job,Go mod check succeeded!)

revive: ## Run revive linter
	$(call before_job,Revive linter is running:)
	revive -exclude vendor/... -formatter friendly -config .revive.toml ./...
	$(call after_job,Revive linter succeeded!)

lint-golangci: ## Runs golangci-lint. It outputs to the code-climate json file in if CI is defined.
	$(call before_job, Running golangci-lint)
ifndef GOLANGCI_LINT
	@echo "Can\'t find executable of the golangci-lint. Make sure it is installed. See github.com\/golangci\/golangci-lint#install"
	@exit 1
endif
	golangci-lint run --max-same-issues 0 --max-issues-per-linter 0 $(if $(CI),--out-format code-climate > gl-code-quality-report.json 2>golangci-stderr-output)
	$(call after_job,Linting with golangci-lint succeeded!)

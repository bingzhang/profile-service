DATE    ?= $(shell date +%FT%T%z)
GOPATH     = $(CURDIR)/.gopath~
BIN      = $(GOPATH)/bin
BASE     = $(GOPATH)/src
DEPBASE  = $(GOPATH)/src/inamkdep
PKGS     = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./...))
BUILDS   = $(or $(BUILD),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list -f "{{if eq .Name \"main\"}}{{.ImportPath}}{{end}}" ./...))

GIT_VERSION=$(shell git describe --match "v*" 2> /dev/null || cat $(CURDIR)/.version 2> /dev/null || echo v0.0-0-)
BASE_VERSION=$(shell echo $(GIT_VERSION) | cut -f1 -d'-')
MAJOR_VERSION=$(shell echo $(BASE_VERSION) | cut -f1 -d'.' | cut -f2 -d'v')
MINOR_VERSION=$(shell echo $(BASE_VERSION) | cut -f2 -d'.')
BUILD_VERSION=$(shell echo $(BASE_VERSION) | cut -f3 -d'.' || echo 0)
BUILD_OFFSET=$(shell echo $(GIT_VERSION) | cut -s -f2 -d'-' )
CODE_OFFSET=$(shell [ -z "$(BUILD_OFFSET)" ] && echo "0" || echo "$(BUILD_OFFSET)")
BUILD_NUMBER=$(shell echo $$(( $(BUILD_VERSION) + $(CODE_OFFSET) )))
VERSION ?= ${MAJOR_VERSION}.${MINOR_VERSION}.${BUILD_NUMBER}

export -n GOBIN
export GOPATH
export PATH=$(GOPATH)/bin: $(shell printenv PATH)

GO      = go
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 25
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")


.PHONY: all
all: fmt lint vendor | $(BASE) ; $(info $(M) building executable(s)… $(VERSION) $(DATE)) @ ## Build program binary
	$Q cd $(CURDIR) && $(GO) generate ./...
	@ret=0 && for d in $(BUILDS); do \
		if expr \"$$d\" : \".gopath~/src\" 1>/dev/null; then SRCPATH=$(CURDIR) ; else SRCPATH=$(CURDIR)/$$d ; fi ;  \
		cd $${SRCPATH} && env GOBIN=$(CURDIR)/bin $(GO) install \
			-tags release \
			-ldflags '-X main.Version=$(VERSION) -X main.Build=$(DATE)' || ret=$$? ; \
	 done ; exit $$ret

$(BASE): ; $(info $(M) setting GOPATH…)
	@mkdir -p $(dir $@)
	@ln -sf .. $@

$(DEPBASE): ; $(info $(M) setting DEPPATH…)
	@mkdir -p $(dir $@)
	@ln -sf . $@

# Tools

$(BIN):
	@mkdir -p $@
	
$(BIN)/%: | $(BIN) $(BASE) ; $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GOPATH=$$tmp go get $(REPOSITORY) && cp $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		rm -rf $$tmp ; exit $$ret

GODEP = $(BIN)/dep
$(GODEP): REPOSITORY=github.com/golang/dep/cmd/dep

GOESC = $(BIN)/esc
$(GOESC): REPOSITORY=github.com/mjibson/esc

GOLINT = $(BIN)/golint
$(GOLINT): REPOSITORY=golang.org/x/lint/golint

# Tests

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt lint vendor | $(BASE) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q cd $(CURDIR) && $(GO) test -gcflags=-l -timeout $(TIMEOUT)s $(ARGS) ./...

.PHONY: cover
cover: fmt lint vendor | $(BASE) ; $(info $(M) running coverage…) @ ## Run code coverage tests
	$Q cd $(BASE) && 2>&1 $(GO) test -gcflags=-l ./... -coverprofile=c.out
	$Q cd $(BASE) && 2>&1 $(GO) tool cover -html=c.out
	$Q cd $(BASE) && 2>&1 rm -f c.out

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q $(GOLINT) -set_exit_status $(PKGS)

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management

vendor: Gopkg.toml | $(BASE) $(DEPBASE) $(GODEP) $(GOESC) ; $(info $(M) retrieving dependencies…)
	$Q cd $(DEPBASE) && $(GODEP) ensure
	@touch $@
.PHONY: vendor-update
vendor-update: vendor | $(BASE) $(DEPBASE) $(GODEP) $(GOESC)
ifeq "$(origin PKG)" "command line"
	$(info $(M) updating $(PKG) dependency…)
	$Q cd $(DEPBASE) && $(GODEP) ensure -update $(PKG)
else
	$(info $(M) updating all dependencies…)
	$Q cd $(DEPBASE) && $(GODEP) ensure -update
endif
	@touch vendor

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(DEPBASE)
	@rm -rf $(GOPATH)
	@rm -rf bin
	@rm -rf vendor
	@rm -f $(CURDIR)/c.out
	@rm -f $(CURDIR)/test.html
	@rm -f $(CURDIR)/Gopkg.lock

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)



# Project Setup
PROJECT_NAME := provider-cloudflare
PROJECT_REPO := github.com/rossigee/$(PROJECT_NAME)

# Set VERSION from VERSION file if it exists and VERSION is not already set
ifeq ($(origin VERSION), undefined)
ifneq (,$(wildcard VERSION))
VERSION := $(shell cat VERSION)
export VERSION
endif
endif

PLATFORMS ?= linux_amd64 linux_arm64
-include build/makelib/common.mk

# Setup Output
-include build/makelib/output.mk

# Setup Go
NPROCS ?= 1
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.Version=$(VERSION)
GO_SUBDIRS += cmd internal apis
GO111MODULE = on
# Override golangci-lint version for modern Go support
GOLANGCILINT_VERSION ?= 2.3.1
-include build/makelib/golang.mk

# Setup Kubernetes tools
UP_VERSION = v0.28.0
UP_CHANNEL = stable
UPTEST_VERSION = v0.11.1
-include build/makelib/k8s_tools.mk

# Setup Images
IMAGES = provider-cloudflare
# Force registry override (can be overridden by make command arguments)
REGISTRY_ORGS = ghcr.io/rossigee
-include build/makelib/imagelight.mk

# Setup XPKG - Standardized registry configuration
# Primary registry: GitHub Container Registry under rossigee
XPKG_REG_ORGS ?= ghcr.io/rossigee
XPKG_REG_ORGS_NO_PROMOTE ?= ghcr.io/rossigee

# Optional registries (can be enabled via environment variables)
# Harbor publishing has been removed - using only ghcr.io/rossigee
# To enable Upbound: export ENABLE_UPBOUND_PUBLISH=true make publish XPKG_REG_ORGS=xpkg.upbound.io/rossigee
XPKGS = provider-cloudflare
-include build/makelib/xpkg.mk

# NOTE: we force image building to happen prior to xpkg build so that we ensure
# image is present in daemon.
xpkg.build.provider-cloudflare: do.build.images

# Setup Package Metadata
CROSSPLANE_VERSION = 1.19.0
-include build/makelib/local.xpkg.mk
-include build/makelib/controlplane.mk

# Targets

# run `make submodules` after cloning the repository for the first time.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# NOTE: the build submodule currently overrides XDG_CACHE_HOME in order to
# force the Helm 3 to use the .work/helm directory. This causes Go on Linux
# machines to use that directory as the build cache as well. We should adjust
# this behavior in the build submodule because it is also causing Linux users
# to duplicate their build cache, but for now we just make it easier to identify
# its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# Use the default generate targets from build system
# The build system already handles code generation properly

# NOTE: we must ensure up is installed in tool cache prior to build as including the k8s_tools
# machinery prior to the xpkg machinery sets UP to point to tool cache.
build.init: $(UP)

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/provider --debug

# NOTE: we ensure up is installed prior to running platform-specific packaging steps in xpkg.build.
xpkg.build: $(UP)

# Ensure CLI is available for package builds and publishing
$(foreach x,$(XPKGS),$(eval xpkg.build.$(x): $(CROSSPLANE_CLI)))

# Rules to build packages for each platform
$(foreach p,$(filter linux_%,$(PLATFORMS)),$(foreach x,$(XPKGS),$(eval $(XPKG_OUTPUT_DIR)/$(p)/$(x)-$(VERSION).xpkg: $(CROSSPLANE_CLI); @$(MAKE) xpkg.build.$(x) PLATFORM=$(p))))

# Ensure packages are built for all platforms before publishing
$(foreach r,$(XPKG_REG_ORGS),$(foreach x,$(XPKGS),$(eval xpkg.release.publish.$(r).$(x): $(CROSSPLANE_CLI) $(foreach p,$(filter linux_%,$(PLATFORMS)),$(XPKG_OUTPUT_DIR)/$(p)/$(x)-$(VERSION).xpkg))))

.PHONY: submodules run

# Additional targets

# Use the default test target from build system
# test: generate
#	@$(INFO) Running tests...
#	@$(GO) test -v ./...

# Run tests with coverage
test.cover: generate
	@$(INFO) Running tests with coverage...
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html

# Install CRDs into a cluster
install-crds: generate
	kubectl apply -f package/crds

# Uninstall CRDs from a cluster
uninstall-crds:
	kubectl delete -f package/crds

# Publish to Upbound Marketplace
publish-upbound: xpkg.build
	@$(INFO) Publishing to Upbound Marketplace...
	@up xpkg push $(PROJECT_REPO)/package/provider-cloudflare-$(VERSION).xpkg xpkg.upbound.io/rossigee/provider-cloudflare:$(VERSION)

# Validate package for marketplace compliance
validate-package:
	@$(INFO) Validating package structure...
	@test -f package/crossplane.yaml || (echo "Missing package/crossplane.yaml" && exit 1)
	@test -f .github/cloudflare-icon.svg || (echo "Missing icon file" && exit 1)
	@grep -q "meta.crossplane.io/maintainer" package/crossplane.yaml || (echo "Missing maintainer annotation" && exit 1)
	@grep -q "meta.crossplane.io/source" package/crossplane.yaml || (echo "Missing source annotation" && exit 1)
	@grep -q "meta.crossplane.io/license" package/crossplane.yaml || (echo "Missing license annotation" && exit 1)
	@$(INFO) Package validation passed!

# Performance benchmarks
benchmark: generate
	@$(INFO) Running performance benchmarks...
	@$(GO) test -bench=. -benchmem ./internal/clients/benchmarks/

# Run benchmarks and save results
benchmark.save: generate
	@$(INFO) Running benchmarks and saving results...
	@mkdir -p .benchmarks
	@$(GO) test -bench=. -benchmem ./internal/clients/benchmarks/ > .benchmarks/benchmark-$(shell date +%Y%m%d-%H%M%S).txt

# Compare benchmark results
benchmark.compare: generate
	@$(INFO) Comparing benchmark results...
	@if [ ! -f .benchmarks/baseline.txt ]; then \
		echo "No baseline found. Run 'make benchmark.baseline' first."; \
		exit 1; \
	fi
	@$(GO) test -bench=. -benchmem ./internal/clients/benchmarks/ > .benchmarks/current.txt
	@benchcmp .benchmarks/baseline.txt .benchmarks/current.txt || echo "benchcmp not installed, showing raw comparison:"
	@echo "Baseline:" && head -10 .benchmarks/baseline.txt
	@echo "Current:" && head -10 .benchmarks/current.txt

# Set current benchmark results as baseline
benchmark.baseline: generate
	@$(INFO) Setting benchmark baseline...
	@mkdir -p .benchmarks
	@$(GO) test -bench=. -benchmem ./internal/clients/benchmarks/ > .benchmarks/baseline.txt
	@$(INFO) Baseline set. Use 'make benchmark.compare' to compare future runs.

# Run benchmarks with CPU profiling
benchmark.profile: generate
	@$(INFO) Running benchmarks with CPU profiling...
	@mkdir -p .benchmarks/profiles
	@$(GO) test -bench=BenchmarkZone -cpuprofile=.benchmarks/profiles/cpu.prof ./internal/clients/benchmarks/
	@$(INFO) Profile saved to .benchmarks/profiles/cpu.prof
	@$(INFO) Analyze with: go tool pprof .benchmarks/profiles/cpu.prof

# Run benchmarks with memory profiling
benchmark.memprofile: generate
	@$(INFO) Running benchmarks with memory profiling...
	@mkdir -p .benchmarks/profiles
	@$(GO) test -bench=BenchmarkZone -memprofile=.benchmarks/profiles/mem.prof ./internal/clients/benchmarks/
	@$(INFO) Profile saved to .benchmarks/profiles/mem.prof
	@$(INFO) Analyze with: go tool pprof .benchmarks/profiles/mem.prof

# Run specific benchmark category
benchmark.zone: generate
	@$(INFO) Running Zone benchmarks...
	@$(GO) test -bench=BenchmarkZone -benchmem ./internal/clients/benchmarks/

benchmark.record: generate
	@$(INFO) Running Record benchmarks...
	@$(GO) test -bench=BenchmarkRecord -benchmem ./internal/clients/benchmarks/

benchmark.loadbalancing: generate
	@$(INFO) Running Load Balancing benchmarks...
	@$(GO) test -bench=BenchmarkLoadBalancer -benchmem ./internal/clients/benchmarks/

benchmark.cache: generate
	@$(INFO) Running Cache benchmarks...
	@$(GO) test -bench=BenchmarkCache -benchmem ./internal/clients/benchmarks/

benchmark.security: generate
	@$(INFO) Running Security benchmarks...
	@$(GO) test -bench=BenchmarkSecurity -benchmem ./internal/clients/benchmarks/

# Extended benchmark run (longer duration for more accurate results)
benchmark.extended: generate
	@$(INFO) Running extended benchmarks (10s each)...
	@$(GO) test -bench=. -benchmem -benchtime=10s ./internal/clients/benchmarks/

# Continuous benchmarking for CI
benchmark.ci: generate
	@$(INFO) Running CI benchmarks...
	@mkdir -p .benchmarks
	@$(GO) test -bench=. -benchmem ./internal/clients/benchmarks/ | tee .benchmarks/ci-$(shell date +%Y%m%d-%H%M%S).txt
	@echo "Benchmark results saved for CI analysis"
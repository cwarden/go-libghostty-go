# The package is pure Go: internal/vt holds the translation of
# libghostty-vt that generator.go produces. Building and testing need only
# the Go toolchain. Regenerating needs Zig 0.17 (see generator.go).

# Keep the exit status of pipelines such as `go run generator.go | tee`.
SHELL = /bin/bash
.SHELLFLAGS = -o pipefail -c

# Work directory for the generator; the ghostty tarball is cached there.
DIR = /tmp/libghostty-generate

# The platforms generated code is committed for.
TARGETS = linux/amd64 linux/arm64

.PHONY: build test bench lint generate generate-all clean

build:
	go build ./...

test:
	go test ./...

# Run every benchmark once; pass BENCH_FLAGS for -count, -benchtime or
# a -bench selector, e.g. `make bench BENCH_FLAGS="-count 5"`.
bench:
	go test -run '^$$' -bench . -benchmem $(BENCH_FLAGS) .

# The generated package is excluded from vet; the unsafe.Pointer check is
# disabled because the translated code passes C pointers as uintptr.
lint:
	go vet -unsafeptr=false . ./sys/... ./examples/... ./internal/libshim

# Translate libghostty-vt for the host GOOS/GOARCH.
generate:
	mkdir -p $(DIR)
	GO_GENERATE_DIR=$(DIR) go run generator.go 2>&1 | tee log-generate

# Translate libghostty-vt for every platform in TARGETS.
generate-all:
	mkdir -p $(DIR)
	for t in $(TARGETS); do \
		GO_GENERATE_GOOS=$${t%/*} GO_GENERATE_GOARCH=$${t#*/} GO_GENERATE_DIR=$(DIR) go run generator.go || exit 1; \
	done

clean:
	rm -rf $(DIR) log-generate
	go clean

# Go Libghostty Bindings

Go bindings for `libghostty-vt`.

This package is pure Go. `libghostty-vt` is compiled into it: Zig's C
backend emits the library as C, and [ccgo](https://modernc.org/ccgo)
translates that C to Go on top of [modernc.org/libc](https://modernc.org/libc).
Building a program that uses it needs only the Go toolchain, and
cross-compiling is `GOOS=... GOARCH=... go build`.

> [!WARNING]
>
> **I'm not promising any API stability yet.** This is a new project and the
> API may change as necessary. The underlying functionality is very stable,
> but the Go API is still being designed.

## Example

```go
package main

import (
	"fmt"
	"log"

	"go.mitchellh.com/libghostty"
)

func main() {
	term, err := libghostty.NewTerminal(libghostty.WithSize(80, 24))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// Feed VT data — bold green "world", then plain text.
	fmt.Fprintf(term, "Hello, \033[1;32mworld\033[0m!\r\n")

	// Format the terminal contents as plain text.
	f, err := libghostty.NewFormatter(term,
		libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
		libghostty.WithFormatterTrim(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	output, _ := f.FormatString()
	fmt.Println(output) // Hello, world!
}
```

More examples are in the [`examples/`](examples/) directory.

## Usage

Add the module to your Go project:

```shell
go get go.mitchellh.com/libghostty
```

Generated code is committed for Linux on amd64 and arm64. On other
platforms the package does not build; generating code for another
platform is described under Regenerating.

## Performance

The translation runs slower than the cgo build did. These are
`go test -bench` results on one machine (Linux, Intel i7-14700HX, Go
1.26), the previous cgo build (commit e247028, libghostty-vt built with
`-Doptimize=ReleaseFast`) against this one; both measured with `-count 3`,
and runs vary by about 20%, so read the ratios as rough.

| Benchmark | cgo | pure Go | ratio |
| --- | ---: | ---: | ---: |
| VTWriteCall (one call, no bytes) | 44 ns | 441 ns | 10× |
| VTWrite/Plain | 251 µs | 1.88 ms | 7× |
| VTWrite/OSCDense | 785 µs | 5.99 ms | 8× |
| FormatterVTWriteTo/Plain80x24 | 10.4 µs | 78 µs | 7× |
| FormatterVTWriteTo/Mixed240x80 | 108 µs | 732 µs | 7× |
| RenderStateUpdate/Text/Clean | 70 ns | 1.10 µs | 16× |
| RenderStateUpdate/Text/Row | 169 ns | 2.93 µs | 17× |
| RenderStateUpdate/Text/Scroll | 903 ns | 4.52 µs | 5× |
| RenderStateRowCellsGetMulti/Keys1 | 156 ns | 273 ns | 1.8× |
| RenderStateRowText/Plain/AppendText | 14.5 µs | 93 µs | 6× |
| RenderStateRowText/Plain/PerCell | 169 µs | 257 µs | 1.5× |
| RenderStateImageFrame/Fixed/Always | 1.49 µs | 6.39 µs | 4× |
| SnapshotEncode/PlainNoScrollback/ReuseBuffer | 2.9 µs | 117 µs | 40× |
| SnapshotEncode/Plain10KScrollback/ReuseBuffer | 415 µs | 32.7 ms | 79× |
| SnapshotDecode/PlainNoScrollback/Full/ZeroCopy | 29 µs | 160 µs | 6× |
| SnapshotDecode/Plain10KScrollback/Full/ZeroCopy | 2.2 ms | 38 ms | 17× |
| geometric mean over all 122 benchmarks | 10.1 µs | 56.9 µs | 5.6× |

Two things account for most of the difference. The translated code is
plain Go with no SIMD: Zig's vectorized scanning of cells and the
hardware CRC32C instruction are replaced by scalar loops and the
table-driven software CRC (a profile of snapshot encoding spends half its
time in `crc32c_Software_updateMulti` and most of the rest in the
per-cell encoding loop). And each translated function allocates its C
stack frame from a per-call TLS stack, which matters for the many small
calls a render state update makes. The bindings themselves add little:
taking a TLS from the pool costs about 10 ns, pinning an input slice about
50 ns, and a getter such as `Terminal.CursorX` costs 23 ns through the
bindings against 6 ns for the translated function alone
(`BenchmarkLibraryCall` and `BenchmarkLibraryCallRaw`).

`make bench` runs every benchmark; `make bench BENCH_FLAGS="-count 5"`
passes extra flags, and `benchstat` compares two result files.

## Development

```shell
make build
make test
```

Both need only Go. The Nix development shell (`nix develop`) also
provides the Zig version the generator needs.

### Regenerating

The translated library lives in `internal/vt`: `vt_<goos>_<goarch>_<n>.go`
(the library, split into files small enough for the Go compiler) and
`consts.go` (the `GHOSTTY_*` constants and the C API types, from the
headers). `generator.go` produces them:

```shell
make generate        # the host platform
make generate-all    # every platform in the Makefile's TARGETS
```

The generator needs Zig 0.17, `curl`, `tar` and `patch`, and network
access: it downloads the ghostty source at the commit pinned in
`generator.go` and Zig fetches the library's dependencies. Linux targets
can be generated on any Linux host because modernc.org/libc ships their C
headers. Regenerate and commit when the pinned commit, the patches, or the
pinned ccgo and libc versions change. Two runs on the same inputs produce
files that differ in the order of the string literal table, which ccgo
assigns in map iteration order.

The generator applies the patches in `internal/patches` to the ghostty
source in file name order:

| Patch | Contents |
| --- | --- |
| `01-link-libc.patch` | Link libc into libghostty-vt so that it allocates with `malloc` instead of `mmap` and raw system calls, which Zig emits as inline assembly. |
| `02-std-zlib.patch` | Decode zlib streams (Kitty graphics compression) with `std.compress.flate` instead of wuffs. The wuffs decoder is written as a coroutine that jumps into loops, which has no translation to Go. The replacement checks the Adler-32 trailer, which the standard decompressor reads without verifying. |

`internal/ccgo` holds what the generator adds to the emitted C: a patch
to Zig's `zig.h` support header, a prelude with the libc headers, and the
128-bit integer helpers of compiler-rt. `internal/libshim` is a Go
package the translation links against for the C functions modernc.org/libc
lacks (pointer-sized atomics, some overflow builtins, `posix_memalign`).

Environment variables honoured by the generator:

| Variable | Effect |
| --- | --- |
| `GO_GENERATE_DIR` | work directory (the Makefile uses /tmp/libghostty-generate); the ghostty tarball is cached there |
| `GO_GENERATE_KEEP` | keep a temporary work directory |
| `GO_GENERATE_GOOS`, `GO_GENERATE_GOARCH` | target platform (default: the host's) |
| `GO_GENERATE_ZIG` | the zig executable (default: `zig` from PATH) |

### Updating libghostty

Change `ghosttyCommit` in `generator.go`, run `make generate-all`, and
adjust the bindings to the new headers, which the generator leaves at
`$GO_GENERATE_DIR/out-<goarch>/include/ghostty/`.

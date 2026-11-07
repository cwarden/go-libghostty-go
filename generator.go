//go:build ignore

// Command generator produces the Go translation of libghostty-vt that the
// package internal/vt is made of.
//
// It downloads the ghostty source at the pinned commit, applies
// internal/patches/*.patch, has Zig's C backend emit libghostty-vt as C
// (zig build -Dofmt=c), rewrites that C for ccgo, translates it with ccgo/v4
// together with the wuffs base module it needs, and writes the result to
// internal/vt/vt_<goos>_<goarch>_<n>.go. It also translates the C headers
// to internal/vt/consts.go, which holds the GHOSTTY_* constants.
//
// Environment:
//
//	GO_GENERATE_DIR     work directory (default: a fresh temporary directory);
//	                    the ghostty tarball is cached there
//	GO_GENERATE_KEEP    keep a temporary work directory when set
//	GO_GENERATE_GOOS    target operating system (default: the host's)
//	GO_GENERATE_GOARCH  target architecture (default: the host's)
//	GO_GENERATE_ZIG     the zig executable (default: zig from PATH)
//
// Zig 0.17 is required. Linux targets can be generated on any Linux host
// because modernc.org/libc ships their C headers.
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	ccgo "modernc.org/ccgo/v4/lib"
)

const (
	ghosttyCommit = "66af4c7458db7e19ae7adb44aa7054879be58b6b"
	ghosttyURL    = "https://github.com/ghostty-org/ghostty/archive/%s.tar.gz"
	modulePath    = "go.mitchellh.com/libghostty"
	outDir        = "internal/vt"
	// The Go compiler stores line numbers in 20 bits, and its checks of
	// goto statements fail in a file with more lines than that.
	maxLines = 400000
)

// zigTargets maps GOOS/GOARCH to the Zig target of the C that is emitted.
// musl is used because modernc.org/libc is derived from musl and declares
// the same symbol names.
var zigTargets = map[string]string{
	"linux/amd64": "x86_64-linux-musl",
	"linux/arm64": "aarch64-linux-musl",
}

// libcFunctions are the libc functions the Zig code declares itself. The
// declarations use Zig's own parameter types (uint32_t flags, for example)
// and the Go linker would reject calls to modernc.org/libc with them, so
// they are dropped and the libc headers in prelude.h declare them instead.
var libcFunctions = []string{
	"write", "madvise", "memcpy", "memmove", "shm_open", "shm_unlink",
	"munmap", "mmap", "readlink", "lseek", "preadv", "close", "unlinkat",
	"realpath", "openat", "readv", "mremap", "malloc", "calloc",
}

// renamedFunctions are routed to internal/libshim, which implements the
// aligned allocation that modernc.org/libc lacks.
var renamedFunctions = []string{"posix_memalign", "free", "realloc", "malloc_usable_size"}

var wuffsDefines = []string{
	"-DWUFFS_CONFIG__MODULES",
	"-DWUFFS_CONFIG__MODULE__BASE",
	"-DWUFFS_IMPLEMENTATION",
	"-DWUFFS_CONFIG__AVOID_CPU_ARCH",
}

var (
	goos   = env("GO_GENERATE_GOOS", runtime.GOOS)
	goarch = env("GO_GENERATE_GOARCH", runtime.GOARCH)
	zig    = env("GO_GENERATE_ZIG", "zig")
)

func env(name, dflt string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return dflt
}

func fail(msg string, args ...any) {
	fmt.Fprintln(os.Stderr, strings.TrimSpace(fmt.Sprintf(msg, args...)))
	os.Exit(1)
}

func must(err error) {
	if err != nil {
		fail("%v", err)
	}
}

func mustRead(path string) string {
	b, err := os.ReadFile(path)
	must(err)
	return string(b)
}

func mustWrite(path, s string) {
	must(os.WriteFile(path, []byte(s), 0o644))
}

// run executes a command in dir, echoing it, and fails on error.
func run(dir string, args ...string) string {
	fmt.Fprintf(os.Stderr, "+ %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		fail("%s: %v", strings.Join(args, " "), err)
	}
	return string(out)
}

func main() {
	target := goos + "/" + goarch
	triple, ok := zigTargets[target]
	if !ok {
		fail("no zig target for %s", target)
	}
	repoRoot, err := os.Getwd()
	must(err)

	dir := os.Getenv("GO_GENERATE_DIR")
	switch {
	case dir != "":
		must(os.MkdirAll(dir, 0o755))
	default:
		dir, err = os.MkdirTemp("", "libghostty-generate-")
		must(err)
		defer func() {
			if os.Getenv("GO_GENERATE_KEEP") == "" {
				os.RemoveAll(dir)
				return
			}
			fmt.Printf("%s: work directory kept\n", dir)
		}()
	}
	dir, err = filepath.Abs(dir)
	must(err)

	src := fetchGhostty(repoRoot, dir)
	prefix := filepath.Join(dir, "out-"+goarch)
	must(os.RemoveAll(prefix))
	run(src, zig, "build", "-Demit-lib-vt", "-Demit-xcframework=false",
		"-Doptimize=ReleaseFast", "-Dsimd=false", "-Dofmt=c",
		"-Dtarget="+triple, "--prefix", prefix)

	work := filepath.Join(dir, "work-"+goarch)
	must(os.RemoveAll(work))
	must(os.MkdirAll(work, 0o755))

	// zig.h is the support header of Zig's C backend.
	zigLib := zigEnv("lib_dir")
	mustWrite(filepath.Join(work, "zig.h"), mustRead(filepath.Join(zigLib, "zig.h")))
	run(work, "patch", "-p0", "-i", filepath.Join(repoRoot, "internal/ccgo/zig.h.patch"))
	mustWrite(filepath.Join(work, "prelude.h"), mustRead(filepath.Join(repoRoot, "internal/ccgo/prelude.h")))

	cSource := rewriteC(mustRead(filepath.Join(prefix, "lib/ghostty-vt.c"))) +
		"\n" + mustRead(filepath.Join(repoRoot, "internal/ccgo/compiler_rt.c"))
	mustWrite(filepath.Join(work, "ghostty-vt.c"), cSource)
	mustWrite(filepath.Join(work, "wuffs-v0.4.c"), mustRead(findWuffs(src)))
	headers := readHeaders(filepath.Join(prefix, "include/ghostty"))
	mustWrite(filepath.Join(work, "consts.c"), "#include <ghostty/vt.h>\n"+macroEnum(headers))

	// ccgo resolves modernc.org/libc and internal/libshim through the Go
	// module system from its working directory. A workspace joining a
	// throwaway module with this repository makes both visible.
	run(work, "go", "mod", "init", "example.com/work")
	run(work, "go", "work", "init", ".", repoRoot)

	ccgoRun(work, "-extended-errors", "-ignore-unsupported-alignment",
		"-U__SIZEOF_FLOAT80__", "-U__SIZEOF_INT128__", "-I.", "-include", "prelude.h",
		"-c", "ghostty-vt.c", "-o", "ghostty-vt.o.go")
	ccgoRun(work, append([]string{"-extended-errors", "-ignore-unsupported-alignment",
		"-U__SIZEOF_INT128__"}, append(wuffsDefines, "-c", "wuffs-v0.4.c", "-o", "wuffs.o.go")...)...)
	ccgoRun(work, "-extended-errors", "-I", filepath.Join(prefix, "include"),
		"-c", "consts.c", "-o", "consts.o.go")

	// ccgo's Go printer emits T{...}.field, which is not valid Go; the
	// parenthesized form is restored after every step that prints Go.
	objPath := filepath.Join(work, "ghostty-vt.o.go")
	mustWrite(objPath, parenthesizeLiterals(mustRead(objPath)))

	ccgoLink(work, "--package-name", "vt", "--prefix-external=X",
		"-L", modulePath+"/internal", "-lshim",
		"-o", "vt.go", "ghostty-vt.o.go", "wuffs.o.go")
	ccgoLink(work, "--package-name", "vt", "-o", "consts.go", "consts.o.go")

	vt := parenthesizeLiterals(mustRead(filepath.Join(work, "vt.go")))
	// The overflow builtins for widths that modernc.org/libc does not
	// provide are in internal/libshim.
	vt = regexp.MustCompile(`\biqlibc\.__builtin_`).ReplaceAllString(vt, "libshim.X__builtin_")

	vt = removeFrameDefers(vt)

	types := newTypeMap(headers, cSource)
	unions := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^union (\w+) \{`).FindAllStringSubmatch(cSource, -1) {
		unions[types.zigToC[m[1]]] = true
		unions[m[1]] = true
	}
	for _, m := range regexp.MustCompile(`(?m)^typedef union \w*\s*\{|^union (\w+) \{`).FindAllStringSubmatch(headers, -1) {
		unions[m[1]] = true
	}
	for _, m := range regexp.MustCompile(`(?s)typedef\s+union\s*\w*\s*\{.*?\}\s*(Ghostty\w+)\s*;`).FindAllStringSubmatch(headers, -1) {
		unions[m[1]] = true
	}
	vt = fixUnions(types.apply(vt), unions)
	consts := fixUnions(types.constsFile(mustRead(filepath.Join(work, "consts.go"))), unions)
	if strings.Contains(consts, "unsafe.Sizeof") {
		consts = strings.Replace(consts, "package vt\n", "package vt\n\nimport \"unsafe\"\n", 1)
	}

	out := filepath.Join(repoRoot, outDir)
	must(os.MkdirAll(out, 0o755))
	old, _ := filepath.Glob(filepath.Join(out, fmt.Sprintf("vt_%s_%s_*.go", goos, goarch)))
	for _, f := range old {
		must(os.Remove(f))
	}
	for i, chunk := range split(vt) {
		writeGo(filepath.Join(out, fmt.Sprintf("vt_%s_%s_%d.go", goos, goarch, i)),
			fmt.Sprintf("//go:build %s && %s\n\n", goos, goarch), chunk)
	}
	writeGo(filepath.Join(out, "consts.go"), "", consts)
	run(repoRoot, "env", "GOOS="+goos, "GOARCH="+goarch, "go", "build", "./"+outDir)
	types.report()
}

// fetchGhostty extracts the pinned ghostty release into dir and applies
// the patches. The tarball is kept in dir for later runs.
func fetchGhostty(repoRoot, dir string) string {
	tar := filepath.Join(dir, "ghostty-"+ghosttyCommit+".tar.gz")
	if _, err := os.Stat(tar); err != nil {
		run(dir, "curl", "-fsSL", "-o", tar, fmt.Sprintf(ghosttyURL, ghosttyCommit))
	}
	src := filepath.Join(dir, "ghostty-"+ghosttyCommit)
	must(os.RemoveAll(src))
	run(dir, "tar", "xzf", tar)
	patches, err := filepath.Glob(filepath.Join(repoRoot, "internal/patches/*.patch"))
	must(err)
	// filepath.Glob returns the names sorted, which is the order the
	// patches apply in.
	for _, p := range patches {
		run(src, "patch", "-p1", "-i", p)
	}
	return src
}

// zigEnv returns a string field of `zig env`, whose output is Zig object
// notation.
func zigEnv(field string) string {
	out := run(".", zig, "env")
	m := regexp.MustCompile(`\.` + field + ` = "([^"]+)"`).FindStringSubmatch(out)
	if m == nil {
		fail("zig env: no %s", field)
	}
	return m[1]
}

// findWuffs locates the wuffs C source that the Zig build fetched, which
// the kitty graphics pixel operations are built from.
func findWuffs(src string) string {
	zon := mustRead(filepath.Join(src, "pkg/wuffs/build.zig.zon"))
	m := regexp.MustCompile(`(?s)\.wuffs = \.\{\s*\.url = "[^"]*",\s*\.hash = "([^"]+)"`).FindStringSubmatch(zon)
	if m == nil {
		fail("pkg/wuffs/build.zig.zon: no wuffs hash")
	}
	candidates := []string{
		filepath.Join(src, "zig-pkg", m[1], "release/c/wuffs-v0.4.c"),
		filepath.Join(zigEnv("global_cache_dir"), "p", m[1], "release/c/wuffs-v0.4.c"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	fail("wuffs source not found; looked in %s", strings.Join(candidates, ", "))
	return ""
}

func ccgoRun(dir string, args ...string) {
	fmt.Fprintf(os.Stderr, "+ ccgo %s\n", strings.Join(args, " "))
	wd, err := os.Getwd()
	must(err)
	must(os.Chdir(dir))
	defer os.Chdir(wd)
	must(ccgo.NewTask(goos, goarch, append([]string{"ccgo"}, args...), os.Stdout, os.Stderr, nil).Main())
}

// ccgoLink links object files. ccgo runs gofmt on its output and reports
// an error when that fails, which happens because of the literals that
// parenthesizeLiterals fixes; the output file is still written and the
// error is accepted when it is the only one.
func ccgoLink(dir string, args ...string) {
	fmt.Fprintf(os.Stderr, "+ ccgo %s\n", strings.Join(args, " "))
	wd, err := os.Getwd()
	must(err)
	must(os.Chdir(dir))
	defer os.Chdir(wd)
	err = ccgo.NewTask(goos, goarch, append([]string{"ccgo"}, args...), os.Stdout, os.Stderr, nil).Main()
	if err == nil {
		return
	}
	out := args[len(args)-1]
	for i, a := range args {
		if a == "-o" {
			out = args[i+1]
		}
	}
	if _, serr := os.Stat(out); serr != nil || !strings.Contains(err.Error(), "gofmt") {
		fail("%v", err)
	}
}

// rewriteC makes the C that Zig emits acceptable to ccgo.
func rewriteC(src string) string {
	// Raw Linux system calls are inline assembly; call libc's __syscall6
	// instead, which returns -errno as the Zig code expects.
	src = regexp.MustCompile(`(?ms)^(static uint64_t os_linux_\w+_syscall(\d)__\d+\([^)]*\)) \{\n.*?^\}\n`).
		ReplaceAllStringFunc(src, func(s string) string {
			m := regexp.MustCompile(`(?ms)^(static uint64_t os_linux_\w+_syscall(\d)__\d+\([^)]*\)) \{\n`).FindStringSubmatch(s)
			n := int(m[2][0] - '0')
			args := ""
			for i := 1; i <= n; i++ {
				args += fmt.Sprintf(", (long)a%d", i)
			}
			args += strings.Repeat(", 0", 6-n)
			return m[1] + " {\n return (uint64_t)__syscall6((long)a0" + args + ");\n}\n"
		})
	// ccgo rejects single-digit hexadecimal character escapes.
	src = regexp.MustCompile(`'\\x([0-9a-fA-F])'`).ReplaceAllString(src, `'\x0$1'`)
	// A compound literal of the software 128-bit integer type inside a
	// brace initializer is miscompiled; the plain brace form is fine.
	src = regexp.MustCompile(`\{zig_make_([iu]128)\(`).ReplaceAllString(src, "{zig_init_$1(")
	// zig_mangled renames a symbol with __asm, which ccgo ignores; use the
	// libc name directly.
	for _, m := range regexp.MustCompile(`zig_mangled\((\w+), "(\w+)"\)`).FindAllStringSubmatch(src, -1) {
		src = strings.ReplaceAll(src, " "+m[0], "")
		src = regexp.MustCompile(`\b`+m[1]+`\b`).ReplaceAllString(src, m[2])
	}
	for _, fn := range libcFunctions {
		src = regexp.MustCompile(`(?m)^zig_extern [^\n]*\b`+fn+`\([^\n]*\);\n`).ReplaceAllString(src, "")
	}
	for _, fn := range renamedFunctions {
		src = regexp.MustCompile(`(^|[^.\w>])`+fn+`\(`).ReplaceAllString(src, "${1}ccgo_"+fn+"(")
	}
	// Zig's C backend leaves locals uninitialized and, for a packed struct
	// literal, sets the fields with read-modify-write operations on the
	// uninitialized storage, so the bits outside the fields keep whatever
	// the storage held. Locals whose address is taken live in the TLS
	// stack frame, which ccgo allocates without zeroing and reuses; those
	// leftover bits then reach hashes and comparisons. A packed struct is
	// a bitpack__ scalar, on its own or inside a struct, so initialize
	// the locals of those types and no others. (Zeroing every frame
	// costs 15-40% on call-heavy paths; initializing every struct local
	// costs about 10% in the formatter.)
	packedTypes := typesWithPackedFields(src)
	src = regexp.MustCompile(`(?m)^( (?:struct|union) (\w+) t\d+) ?;$`).ReplaceAllStringFunc(src, func(l string) string {
		m := regexp.MustCompile(`^( (?:struct|union) (\w+) t\d+) ?;$`).FindStringSubmatch(l)
		if packedTypes[m[2]] {
			return m[1] + " = {0};"
		}
		return l
	})
	src = regexp.MustCompile(`(?m)^( bitpack__\w+ t\d+) ?;$`).ReplaceAllString(src, "$1 = 0;")
	src = dropLayoutAsserts(src)
	return src
}

// dropLayoutAsserts removes the size and alignment assertions for types
// that ccgo lays out differently: packed structs (ccgo ignores the packed
// attribute), alignments above 8 and long double. None of these types
// cross the C API.
func dropLayoutAsserts(src string) string {
	packed := map[string]bool{}
	for _, m := range regexp.MustCompile(`zig_packed\(struct (\w+) \{`).FindAllStringSubmatch(src, -1) {
		packed[m[1]] = true
	}
	// Structs that contain a packed struct by value have the same
	// problem, so propagate until nothing changes.
	defs := regexp.MustCompile(`(?ms)^struct (\w+) \{(.*?)\};`).FindAllStringSubmatch(src, -1)
	for changed := true; changed; {
		changed = false
		for _, d := range defs {
			if packed[d[1]] {
				continue
			}
			for p := range packed {
				if regexp.MustCompile(`\bstruct ` + p + `\b[^*]`).MatchString(d[2]) {
					packed[d[1]] = true
					changed = true
					break
				}
			}
		}
	}
	assert := regexp.MustCompile(`zig_alignOf\((.*)\) == (\d+), "abi mismatch"\);$`)
	lines := strings.Split(src, "\n")
	out := lines[:0]
	for _, l := range lines {
		if strings.HasPrefix(l, "zig_static_assert(") {
			if m := assert.FindStringSubmatch(l); m != nil {
				if m[2] != "1" && m[2] != "2" && m[2] != "4" && m[2] != "8" ||
					strings.Contains(m[1], "long double") || mentionsAny(m[1], packed) {
					continue
				}
			}
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func mentionsAny(s string, names map[string]bool) bool {
	for _, w := range regexp.MustCompile(`\w+`).FindAllString(s, -1) {
		if names[w] {
			return true
		}
	}
	return false
}

// parenthesizeLiterals rewrites T{...}.field to (T{...}).field. ccgo's Go
// printer emits the former, which the Go parser rejects.
func parenthesizeLiterals(src string) string {
	start := regexp.MustCompile(`(^|[^.A-Za-z0-9_])([A-Za-z_][A-Za-z0-9_]*)\{`)
	var out strings.Builder
	i := 0
	for {
		loc := start.FindStringSubmatchIndex(src[i:])
		if loc == nil {
			out.WriteString(src[i:])
			return out.String()
		}
		s := i + loc[4] // the identifier
		j := i + loc[1] // after the brace
		depth := 1
		for depth > 0 {
			switch src[j] {
			case '{':
				depth++
			case '}':
				depth--
			case '"', '\'':
				q := src[j]
				for j++; src[j] != q; j++ {
					if src[j] == '\\' {
						j++
					}
				}
			case '`':
				for j++; src[j] != '`'; j++ {
				}
			}
			j++
		}
		if j < len(src) && src[j] == '.' && src[s-1] != '(' {
			out.WriteString(src[i:s])
			out.WriteString("(")
			out.WriteString(src[s:j])
			out.WriteString(")")
		} else {
			out.WriteString(src[i:j])
		}
		i = j
	}
}

// split cuts a Go file into chunks of at most maxLines lines at function
// boundaries, each with the imports it uses.
func split(src string) []string {
	lines := strings.Split(src, "\n")
	end := 0
	for i, l := range lines {
		if l == ")" {
			end = i
			break
		}
	}
	var imports []string
	var pkg string
	for _, l := range lines[:end] {
		if strings.HasPrefix(l, "\t\"") {
			imports = append(imports, strings.TrimSpace(l))
		}
		if strings.HasPrefix(l, "package ") {
			pkg = l
		}
	}
	var chunks [][]string
	var cur []string
	for _, l := range lines[end+1:] {
		if len(cur) >= maxLines && strings.HasPrefix(l, "func ") {
			chunks = append(chunks, cur)
			cur = nil
		}
		cur = append(cur, l)
	}
	chunks = append(chunks, cur)
	comment := regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*`)
	var out []string
	for _, c := range chunks {
		text := strings.Join(c, "\n")
		code := comment.ReplaceAllString(text, "")
		var b strings.Builder
		b.WriteString(pkg + "\n\nimport (\n")
		for _, imp := range imports {
			name := imp[strings.LastIndex(imp, "/")+1:]
			name = strings.Trim(name, `"`)
			if regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\.`).MatchString(code) {
				b.WriteString("\t" + imp + "\n")
			}
		}
		b.WriteString(")\n\n")
		b.WriteString(text)
		out = append(out, b.String())
	}
	return out
}

func writeGo(path, buildTag, src string) {
	src = "// Code generated by generator.go. DO NOT EDIT.\n\n" + buildTag + src
	b, err := format.Source([]byte(src))
	if err != nil {
		mustWrite(path, src)
		fail("%s: %v", path, err)
	}
	must(os.WriteFile(path, b, 0o644))
}

// ctype is a C type as it appears in a prototype or a field: its base
// name with pointer depth.
type ctype struct {
	base  string
	depth int
}

var typeNoise = regexp.MustCompile(`\b(const|volatile|struct|union|enum)\b`)

// parseCType reads a parameter or field declaration such as
// "const GhosttyAllocator* allocator" or "struct X const *a0".
func parseCType(s string) ctype {
	s = typeNoise.ReplaceAllString(s, "")
	s = regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(s, "")
	depth := strings.Count(s, "*")
	s = strings.ReplaceAll(s, "*", " ")
	f := strings.Fields(s)
	if len(f) == 0 {
		return ctype{}
	}
	return ctype{f[0], depth}
}

func parseParams(s string) []ctype {
	s = strings.TrimSpace(s)
	if s == "" || s == "void" {
		return nil
	}
	var out []ctype
	for _, p := range strings.Split(s, ",") {
		out = append(out, parseCType(p))
	}
	return out
}

type proto struct {
	ret    ctype
	params []ctype
}

type cstruct struct {
	fields []ctype
}

// typeMap renames the types Zig numbers (struct terminal_c_io_Writer_7090)
// to the names the C headers give them (GhosttyWriter), found by matching
// the prototypes of the ghostty_* functions parameter by parameter, and
// then the fields of matched structs.
type typeMap struct {
	zigToC   map[string]string
	opaque   map[string]bool
	unmapped map[string]bool
}

// readHeaders returns the text of every header under includeDir with the
// comments removed.
func readHeaders(includeDir string) string {
	comment := regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*`)
	var b strings.Builder
	must(filepath.Walk(includeDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(path, ".h") {
			b.WriteString(comment.ReplaceAllString(mustRead(path), ""))
			b.WriteString("\n")
		}
		return err
	}))
	return b.String()
}

// macroPrefix marks the enumerators that macroEnum derives from macros.
const macroPrefix = "ccgo_macro_"

// macroEnum returns a C enum with one enumerator per integer-valued
// GHOSTTY_* macro, so that ccgo emits the macro's value as a constant.
// On its own, ccgo emits only the macros whose body is a single literal,
// and emits a suffixed literal such as 0xFFFu as a string. A macro whose
// body calls a function or names anything but numbers, other macros,
// Ghostty types and the integer limits is left out.
func macroEnum(headers string) string {
	number := regexp.MustCompile(`\b(0[xX][0-9a-fA-F]+|\d+)[uUlL]*\b`)
	ident := regexp.MustCompile(`\b[A-Za-z_]\w*\b`)
	allowed := regexp.MustCompile(`^(GHOSTTY_\w+|Ghostty\w+|U?INT\w*_(MAX|MIN)|SIZE_MAX)$`)
	seen := map[string]bool{}
	var b strings.Builder
	b.WriteString("enum {\n")
	// A macro body may continue over several lines.
	headers = strings.ReplaceAll(headers, "\\\n", " ")
	for _, m := range regexp.MustCompile(`(?m)^#define (GHOSTTY_\w+)[ \t]+(.*)$`).FindAllStringSubmatch(headers, -1) {
		name, value := m[1], strings.TrimSpace(m[2])
		if seen[name] || value == "" || strings.ContainsAny(value, "\"'\\") {
			continue
		}
		ok := true
		for _, id := range ident.FindAllString(number.ReplaceAllString(value, " "), -1) {
			if !allowed.MatchString(id) {
				ok = false
			}
		}
		if !ok {
			continue
		}
		seen[name] = true
		fmt.Fprintf(&b, "\t%s%s = (%s),\n", macroPrefix, name, name)
	}
	b.WriteString("};\n")
	return b.String()
}

func newTypeMap(headers, zigC string) *typeMap {
	t := &typeMap{zigToC: map[string]string{}, opaque: map[string]bool{}, unmapped: map[string]bool{}}
	for _, m := range regexp.MustCompile(`typedef\s+(?:const\s+)?struct\s+\w+\s*\*\s*(Ghostty\w+)\s*;`).FindAllStringSubmatch(headers, -1) {
		t.opaque[m[1]] = true
	}

	cProtos := map[string]proto{}
	for _, m := range regexp.MustCompile(`(?s)GHOSTTY_API\s+([^;{}()]*?)\s*\b(ghostty_\w+)\s*\(([^;{}]*?)\)\s*;`).FindAllStringSubmatch(headers, -1) {
		cProtos[m[2]] = proto{parseCType(m[1]), parseParams(m[3])}
	}
	zigProtos := map[string]proto{}
	for _, m := range regexp.MustCompile(`(?m)^zig_extern\s+(.+?)\s*\b(ghostty_\w+)\((.*)\);$`).FindAllStringSubmatch(zigC, -1) {
		zigProtos[m[2]] = proto{parseCType(m[1]), parseParams(m[3])}
	}

	cStructs := map[string]cstruct{}
	for _, m := range regexp.MustCompile(`(?s)typedef\s+(?:struct|union)\s*\w*\s*\{(.*?)\}\s*(Ghostty\w+)\s*;`).FindAllStringSubmatch(headers, -1) {
		cStructs[m[2]] = parseFields(m[1])
	}
	zigStructs := map[string]cstruct{}
	for _, m := range regexp.MustCompile(`(?ms)^(?:zig_packed\()?(?:struct|union) (\w+) \{(.*?)^\}\)?;`).FindAllStringSubmatch(zigC, -1) {
		zigStructs[m[1]] = parseFields(regexp.MustCompile(`zig_(?:under_)?align\(\d+\)`).ReplaceAllString(m[2], ""))
	}

	matched := 0
	cToZig := map[string]string{}
	var pending []string
	add := func(z, c ctype) {
		if !strings.HasPrefix(c.base, "Ghostty") || t.opaque[c.base] || c.depth != z.depth {
			return
		}
		if !regexp.MustCompile(`_\d+$`).MatchString(z.base) {
			t.unmapped[c.base+" <- "+z.base] = true
			return
		}
		// Zig passes a pointer to an array where the header has a pointer
		// to its first element; the element type stays as the header
		// defines it.
		if strings.HasPrefix(z.base, "arr_") {
			return
		}
		if prev, ok := t.zigToC[z.base]; ok {
			if prev != c.base {
				fmt.Fprintf(os.Stderr, "type %s maps to both %s and %s\n", z.base, prev, c.base)
			}
			return
		}
		// Zig has, for example, both its own key enum and the C ABI one;
		// the first Zig type keeps the header name and the others keep
		// theirs. ccgo declares them as aliases of the same underlying
		// type, so the Go signatures still agree.
		if prev, ok := cToZig[c.base]; ok {
			fmt.Fprintf(os.Stderr, "type %s is %s; %s keeps its name\n", c.base, prev, z.base)
			return
		}
		cToZig[c.base] = z.base
		t.zigToC[z.base] = c.base
		pending = append(pending, z.base)
	}
	names := make([]string, 0, len(cProtos))
	for name := range cProtos {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := cProtos[name]
		z, ok := zigProtos[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "%s: declared in the headers but not emitted\n", name)
			continue
		}
		if len(c.params) != len(z.params) {
			fmt.Fprintf(os.Stderr, "%s: %d parameters in the header, %d emitted\n", name, len(c.params), len(z.params))
			continue
		}
		matched++
		add(z.ret, c.ret)
		for i := range c.params {
			add(z.params[i], c.params[i])
		}
	}
	for len(pending) > 0 {
		z := pending[0]
		pending = pending[1:]
		zs, ok1 := zigStructs[z]
		cs, ok2 := cStructs[t.zigToC[z]]
		if !ok1 || !ok2 {
			continue
		}
		if len(zs.fields) != len(cs.fields) {
			fmt.Fprintf(os.Stderr, "%s: %d fields in the header, %d in %s\n", t.zigToC[z], len(cs.fields), len(zs.fields), z)
			continue
		}
		for i := range cs.fields {
			add(zs.fields[i], cs.fields[i])
		}
	}
	fmt.Fprintf(os.Stderr, "matched %d of %d API functions, mapped %d types\n", matched, len(cProtos), len(t.zigToC))
	return t
}

// parseFields reads the fields of a struct body. Function pointer fields
// are kept as placeholders so positions still line up.
func parseFields(body string) cstruct {
	var s cstruct
	for _, f := range strings.Split(body, ";") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if strings.Contains(f, "(") {
			s.fields = append(s.fields, ctype{})
			continue
		}
		s.fields = append(s.fields, parseCType(f))
	}
	return s
}

func (t *typeMap) apply(src string) string {
	if len(t.zigToC) == 0 {
		return src
	}
	names := make([]string, 0, len(t.zigToC))
	for z := range t.zigToC {
		names = append(names, regexp.QuoteMeta(z))
	}
	sort.Strings(names)
	re := regexp.MustCompile(`\b(` + strings.Join(names, "|") + `)\b`)
	return re.ReplaceAllStringFunc(src, func(s string) string { return t.zigToC[s] })
}

// constsFile reduces the translated headers to the GHOSTTY_* constants
// and the Ghostty* types that no renamed Zig type provides: the opaque
// handles and the types nothing in the API prototypes mentions. The
// translated headers also declare what limits.h and stdint.h define, which
// the translated library declares as well.
func (t *typeMap) constsFile(src string) string {
	mapped := map[string]bool{}
	for _, c := range t.zigToC {
		mapped[c] = true
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "consts.go", src, 0)
	must(err)
	// A macro evaluated through macroEnum takes the place of the constant
	// ccgo emitted for the macro itself.
	fromMacro := map[string]bool{}
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.CONST {
			for _, spec := range gd.Specs {
				if name := spec.(*ast.ValueSpec).Names[0].Name; strings.HasPrefix(name, macroPrefix) {
					fromMacro[strings.TrimPrefix(name, macroPrefix)] = true
				}
			}
		}
	}
	var b strings.Builder
	b.WriteString("package vt\n\n")
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST && gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			var name string
			switch s := spec.(type) {
			case *ast.ValueSpec:
				name = s.Names[0].Name
				if strings.HasPrefix(name, macroPrefix) {
					name = strings.TrimPrefix(name, macroPrefix)
					s.Names[0].Name = name
				} else if fromMacro[name] {
					continue
				}
			case *ast.TypeSpec:
				name = s.Name.Name
			}
			if !strings.HasPrefix(name, "GHOSTTY_") && !strings.HasPrefix(name, "Ghostty") || mapped[name] {
				continue
			}
			one := &ast.GenDecl{Tok: gd.Tok, Specs: []ast.Spec{spec}}
			must(printer.Fprint(&b, fset, one))
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

func (t *typeMap) report() {
	if len(t.unmapped) == 0 {
		return
	}
	names := make([]string, 0, len(t.unmapped))
	for n := range t.unmapped {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Fprintf(os.Stderr, "header types without a Zig counterpart in the API: %s\n", strings.Join(names, ", "))
}

// fixUnions changes how the translation stores a C union. ccgo emits a
// union as a struct whose members are zero-length arrays at offset 0, plus
// one member kept as a real field and a byte array of padding up to the
// union's size. The bytes inside the real member's own trailing padding
// belong to the other members, but a Go struct copy need not preserve
// padding, so a union passed or assigned by value could lose them (Zig's
// tagged unions are copied by value all the time). The real member and
// the padding become one byte array of the union's size, which copies
// preserve. The zero-length members stay; they keep the alignment and
// give the address a type. Code that named the real member must read it
// through the union's address instead.
//
// unions holds the names of the C union types; a named Go struct is
// rewritten only when it is one of them (Zig also emits zero-length
// _padding arrays in ordinary structs). A nested anonymous struct block
// is always a union: Zig names every struct type, so the only anonymous
// blocks it emits are the payload unions of its tagged unions.
func fixUnions(src string, unions map[string]bool) string {
	var out strings.Builder
	fixUnionBlocks(src, &out, unions, true)
	return rewriteUnionStores(fixInlineUnions(out.String()), unions)
}

// fixInlineUnions rewrites the unions that ccgo spells inline in
// expressions, as `struct{Fa [0]T; Fb U; F__ccgo_pad1 [n]byte}`, so that
// they stay identical to the declared types. The fields of an inline
// struct are separated by "; "; nested inline structs are rewritten first.
func fixInlineUnions(src string) string {
	var out strings.Builder
	for {
		i := strings.Index(src, "struct{")
		if i < 0 {
			out.WriteString(src)
			return out.String()
		}
		out.WriteString(src[:i])
		src = src[i+len("struct{"):]
		depth, j := 1, 0
		for depth > 0 {
			k := strings.IndexAny(src[j:], "{}")
			if k < 0 {
				fail("fixInlineUnions: unbalanced braces")
			}
			j += k + 1
			if src[j-1] == '{' {
				depth++
			} else {
				depth--
			}
		}
		body := fixInlineUnions(src[:j-1])
		out.WriteString("struct{")
		out.WriteString(rewriteInlineUnionBody(body))
		out.WriteString("}")
		src = src[j:]
	}
}

// rewriteInlineUnionBody applies the union rule to the "; "-separated
// fields of an inline struct.
func rewriteInlineUnionBody(body string) string {
	// Split at depth 0 only; a nested struct field holds "; " itself.
	var fields []string
	depth, start := 0, 0
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '{':
			depth++
		case '}':
			depth--
		case ';':
			if depth == 0 && i+1 < len(body) && body[i+1] == ' ' {
				fields = append(fields, body[start:i])
				start = i + 2
				i++
			}
		}
	}
	fields = append(fields, body[start:])
	var zero []string
	var real, pad string
	for _, f := range fields {
		name, typ, ok := strings.Cut(f, " ")
		if !ok {
			return body
		}
		switch {
		case strings.HasPrefix(typ, "[0]"):
			zero = append(zero, f)
		case name == "F__ccgo_union", name == "F__ccgo_union_align":
			return body
		case strings.HasPrefix(name, "F__ccgo_pad"):
			pad = regexp.MustCompile(`\[(\d+)\]byte`).FindStringSubmatch(typ)[1]
		case real != "":
			return body
		default:
			real = typ
		}
	}
	if len(zero) == 0 || real == "" {
		return body
	}
	size := "unsafe.Sizeof(*new(" + real + "))"
	if pad != "" {
		size += " + " + pad
	}
	return strings.Join(append(zero, "F__ccgo_union_align [0]"+real, "F__ccgo_union ["+size+"]byte"), "; ")
}

var unionField = regexp.MustCompile(`^(\t+)(\w+)\s+(.*)$`)

// fixUnionBlocks copies src to out, rewriting every `struct {` block
// (innermost first) whose direct fields include a zero-length array.
func fixUnionBlocks(src string, out *strings.Builder, unions map[string]bool, top bool) {
	for {
		i := strings.Index(src, "struct {\n")
		if i < 0 {
			out.WriteString(src)
			return
		}
		// At the top level the block belongs to a `type X = struct {`
		// declaration; nested blocks are the anonymous unions.
		// A type declaration is a union only when the C declares it so;
		// a block spelled inside an expression is a union when its
		// fields say so (ccgo spells union values with their type).
		isUnion := true
		if top {
			line := src[strings.LastIndex(src[:i], "\n")+1 : i]
			if m := regexp.MustCompile(`^type (\w+) = $`).FindStringSubmatch(line); m != nil {
				isUnion = unions[m[1]]
			}
		}
		out.WriteString(src[:i+len("struct {\n")])
		src = src[i+len("struct {\n"):]
		// Find the end of this block; the body may hold nested blocks.
		depth, j := 1, 0
		for depth > 0 {
			k := strings.IndexAny(src[j:], "{}")
			if k < 0 {
				fail("fixUnions: unbalanced braces")
			}
			j += k + 1
			if src[j-1] == '{' {
				depth++
			} else {
				depth--
			}
		}
		// j is just past the closing brace; the line holds "<indent>}".
		end := strings.LastIndex(src[:j-1], "\n") + 1
		var body strings.Builder
		fixUnionBlocks(src[:end], &body, unions, false)
		if isUnion {
			out.WriteString(rewriteUnionBody(body.String()))
		} else {
			out.WriteString(body.String())
		}
		out.WriteString(src[end:j])
		src = src[j:]
	}
}

// rewriteUnionBody rewrites the direct fields of one struct body when they
// describe a union. Nested blocks inside the body are already rewritten and
// are skipped by indentation: a direct field starts at the body's own depth.
func rewriteUnionBody(body string) string {
	lines := strings.Split(body, "\n")
	indent := -1
	isUnion := false
	for _, l := range lines {
		m := unionField.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if indent < 0 {
			indent = len(m[1])
		}
		if len(m[1]) == indent && strings.HasPrefix(m[3], "[0]") {
			isUnion = true
		}
	}
	if !isUnion {
		return body
	}
	var out []string
	var real, pad string
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		m := unionField.FindStringSubmatch(l)
		if m == nil || len(m[1]) != indent || strings.HasPrefix(m[3], "[0]") {
			out = append(out, l)
			continue
		}
		if strings.HasPrefix(m[2], "F__ccgo_pad") {
			pad = regexp.MustCompile(`\[(\d+)\]byte`).FindStringSubmatch(m[3])[1]
			continue
		}
		if m[2] == "F__ccgo_union" || m[2] == "F__ccgo_union_align" {
			return body
		}
		// Zig emits zero-length _padding arrays in ordinary structs as
		// well; a union has exactly one real member.
		if real != "" {
			return body
		}
		real = m[3]
		// A real member of anonymous struct type spans several lines.
		if strings.HasSuffix(real, "{") {
			depth := 1
			for depth > 0 {
				i++
				real += "\n" + lines[i]
				depth += strings.Count(lines[i], "{") - strings.Count(lines[i], "}")
			}
		}
	}
	if real == "" {
		return body
	}
	size := "unsafe.Sizeof(*new(" + real + "))"
	if pad != "" {
		size += " + " + pad
	}
	// A byte array has alignment 1; a zero-length array of the real
	// member's type keeps the union's alignment, and with it the offsets
	// of the fields after it, which the translation computes itself.
	out = append(out, strings.Repeat("\t", indent)+"F__ccgo_union_align [0]"+real)
	out = append(out, strings.Repeat("\t", indent)+"F__ccgo_union ["+size+"]byte")
	return strings.Join(out, "\n")
}

// rewriteUnionStores rewrites `tN.Fmember = tM`, a store to the real
// member of a local union variable, as a copy of tM's bytes into the
// union's storage. The three such stores in the translation all have this
// shape; anything else that names a union member fails to compile.
func rewriteUnionStores(src string, unions map[string]bool) string {
	store := regexp.MustCompile(`^(\t+)(t\d+)\.(F\w+) = (t\d+)$`)
	lines := strings.Split(src, "\n")
	funcStart := 0
	for i, l := range lines {
		if strings.HasPrefix(l, "func ") {
			funcStart = i
			continue
		}
		m := store.FindStringSubmatch(l)
		if m == nil || m[3] == "F__ccgo_union" {
			continue
		}
		decl := regexp.MustCompile(`^\tvar (?:\w+, )*` + m[2] + `(?:, \w+)* (\S+)$`)
		for j := funcStart; j < i; j++ {
			if d := decl.FindStringSubmatch(lines[j]); d != nil && unions[d[1]] {
				size := "unsafe.Sizeof(" + m[4] + ")"
				lines[i] = m[1] + "*(*[" + size + "]byte)(unsafe.Pointer(&" + m[2] + ")) = *(*[" + size + "]byte)(unsafe.Pointer(&" + m[4] + "))"
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

// removeFrameDefers replaces the `defer tls.Free(n)` that releases a
// function's C stack frame with a call before each return. The bodies
// are full of goto statements, which keep Go from open-coding a defer,
// so every call paid a heap-allocated defer record; that was a third of
// the cost of a call that does no work. Every translated function names
// its result (r, or another name when a C parameter is called r), so
// `return expr` becomes an assignment to it, the free, and a bare
// return. A return expression that spans lines (a composite literal) is
// followed to its closing brace.
func removeFrameDefers(src string) string {
	lines := strings.Split(src, "\n")
	var out []string
	free, result := "", ""
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if strings.HasPrefix(l, "func ") {
			free = ""
			// The result is named r unless a parameter took that name
			// (C code may have one), so read it from the signature.
			result = ""
			if m := regexp.MustCompile(`\) \((\w+) [^()]+\) \{$`).FindStringSubmatch(l); m != nil {
				result = m[1]
			}
		}
		if m := regexp.MustCompile(`^\tdefer tls\.Free\((\d+)\)$`).FindStringSubmatch(l); m != nil {
			free = "tls.Free(" + m[1] + ")"
			continue
		}
		if free == "" {
			out = append(out, l)
			continue
		}
		if l == "}" {
			// The end of a function; a void function may fall off
			// without a return, and then the frame is freed here. After
			// a top-level return the free would be unreachable, and a
			// function with a result must end in a terminating statement.
			if len(out) == 0 || out[len(out)-1] != "\treturn" {
				out = append(out, "\t"+free)
			}
			out = append(out, l)
			free = ""
			continue
		}
		m := regexp.MustCompile(`^(\t+)return(?: (.*))?$`).FindStringSubmatch(l)
		if m == nil {
			out = append(out, l)
			continue
		}
		indent := m[1]
		if m[2] == "" {
			out = append(out, indent+free, l)
			continue
		}
		if result == "" {
			fail("removeFrameDefers: return with a value in a function without a named result: %s", l)
		}
		out = append(out, indent+result+" = "+m[2])
		depth := strings.Count(m[2], "{") - strings.Count(m[2], "}")
		for depth > 0 {
			i++
			out = append(out, lines[i])
			depth += strings.Count(lines[i], "{") - strings.Count(lines[i], "}")
		}
		out = append(out, indent+free, indent+"return")
	}
	return strings.Join(out, "\n")
}

// typesWithPackedFields returns the names of the struct and union types
// that hold a packed struct (a bitpack__ field), directly or through a
// nested struct or array member.
func typesWithPackedFields(src string) map[string]bool {
	defs := map[string]string{}
	for _, m := range regexp.MustCompile(`(?ms)^(?:zig_packed\()?(?:struct|union) (\w+) \{(.*?)^\}\)?;`).FindAllStringSubmatch(src, -1) {
		defs[m[1]] = m[2]
	}
	packed := map[string]bool{}
	for name, body := range defs {
		if strings.Contains(body, "bitpack__") {
			packed[name] = true
		}
	}
	member := regexp.MustCompile(`\b(?:struct|union) (\w+)\b[^*]`)
	for changed := true; changed; {
		changed = false
		for name, body := range defs {
			if packed[name] {
				continue
			}
			for _, m := range member.FindAllStringSubmatch(body, -1) {
				if packed[m[1]] {
					packed[name] = true
					changed = true
					break
				}
			}
		}
	}
	return packed
}

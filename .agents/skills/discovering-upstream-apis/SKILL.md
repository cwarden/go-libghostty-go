---
name: discovering-upstream-apis
description: >-
  Discovers new or requested C APIs in upstream libghostty-vt headers
  and adds Go bindings. Use when asked to add new APIs, add bindings,
  check for new upstream APIs, or bind a specific API like 'add the
  decode png api'.
---

# Discovering Upstream APIs

Finds new or requested C APIs in the upstream libghostty-vt headers
and creates or updates Go bindings for them.

## Checking for upstream changes

The pinned commit is the `ghosttyCommit` constant in `generator.go`.

Run `scripts/latest-upstream-commit.sh` to compare the pinned commit
against the latest upstream `main` branch. It prints the pinned SHA,
the latest SHA, and whether an update is available.

When an update is available:

1. Update `ghosttyCommit` in `generator.go` to the latest commit.
2. Run `make generate-all` to fetch the new source and regenerate
   `internal/vt`.
3. Run the discovery steps above.
4. Compare against the current Go bindings.
5. Present findings to the user.

## Workflow

### 1. Identify what to bind

Determine the scope based on the user's request:

- **Specific API** (e.g. "add the decode png api"): Search the
  upstream headers for matching functions/types.
- **All new APIs** (e.g. "add the new apis"): Diff the upstream
  headers against existing Go bindings to find unbound APIs.

### 2. Locate upstream headers

Headers live at:

```
/tmp/libghostty-generate/out-amd64/include/ghostty/vt/
```

The umbrella header is at
`/tmp/libghostty-generate/out-amd64/include/ghostty/vt.h` and includes
all sub-headers. Individual headers are in the `vt/` subdirectory.

If that directory does not exist, run `make generate` first.

### 3. Discover unbound APIs

To find what's new or missing:

1. List all C functions in the upstream headers:

   ```
   grep -rh '^[A-Za-z].*ghostty_' \
     /tmp/libghostty-generate/out-amd64/include/ghostty/vt/ \
     | grep '('
   ```

2. List all C functions already referenced in Go files:

   ```
   grep -rh 'C\.\(ghostty_[a-z_]*\)' *.go \
     | grep -oE 'ghostty_[a-z_]+' | sort -u
   ```

3. Cross-reference to find unbound functions.
4. Also check `TODO.md` for known missing items.

For type/enum discovery:

```
grep -rh 'typedef\|^} Ghostty\|GHOSTTY_[A-Z_]*[, ]' \
  /tmp/libghostty-generate/out-amd64/include/ghostty/vt/*.h
```

### 4. Confirm with the user

Before writing code, present the list of new APIs found and ask
which ones the user wants bound. Group them by header file. Include:

- Function signatures
- Related types/enums they depend on
- Which header they come from

### 5. Write Go bindings

Follow the conventions specified in AGENTS.md and patterns in
existing code. Here are some examples:

- **Simple getters** (`ghostty_terminal_get`): See
  `terminal_data.go` — call `ghostty_terminal_get` with the
  appropriate enum, cast result to Go type.
- **New/Free lifecycle**: See `render_state.go` — `NewX()` returns
  `(*X, error)`, `Close()` frees.
- **Effect callbacks**: See `terminal_effect.go` — top-level Go
  trampolines registered with `vt.Fp` and a `handle` (ccgo.go) as
  userdata; the conventions are in `CCGO.md`.
- **Tagged unions**: See `terminal.go` `ScrollViewport*` — set tag
  then poke the value union via `unsafe.Pointer`.
- **Formatters/iterators**: See `formatter.go` — functional options
  pattern, alloc + copy + free for output buffers.

### 6. Write tests

- Add tests in a `_test.go` file matching the source file name.
- Follow existing test patterns (see `terminal_test.go`,
  `formatter_test.go`).
- Run `make test` to verify.

### 7. Update TODO.md

- Remove any items from `TODO.md` that have been bound.
- Add any new partial items if applicable.

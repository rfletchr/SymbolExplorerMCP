# Symbol Explorer MCP

When an LLM needs to understand an unfamiliar codebase, the naive approach is to read files. That burns tokens fast: a medium-sized project can easily exceed a context window before the model has found what it is looking for, and most of what gets read is irrelevant.

Symbol Explorer solves this by giving the LLM a structured index instead of raw source. It parses files with tree-sitter and extracts named declarations - functions, types, methods, constants - with their signatures, doc strings, and import lists, without returning function bodies. The LLM can orient itself in a large project with a single tool call, drill into the files and symbols that matter, and only read full source when it has a specific reason to.

It ships as a Go library, a CLI tool (`symex`), and an MCP server (`symex-mcp`) with six navigation tools designed around the way LLMs actually explore code.

## MCP Tools

| Tool | Description |
|------|-------------|
| `directory_summary` | One line per subdirectory with file + symbol counts. Start here. |
| `file_summary` | File-level symbol counts, paged (`offset`/`limit`). Supports `glob`. |
| `index` | Symbol list grouped by file. `detail: outline` (default) or `full`. |
| `find_symbols` | Regex search on symbol names. Optional `kind` filter. |
| `read_symbol` | Fetch a symbol's signature or full body (`detail: full`). |
| `find_importers` | Find files that import a package, matched by regex. |

All tools accept `root` (required), `use_gitignore`, and `glob` parameters.

## Languages

Go, Python, Rust, C, C++, TypeScript, JavaScript

## Building

```sh
make build          # current platform, outputs to dist/
make dist-linux     # static Linux binaries via zig cc (any host with zig in PATH)
make dist-darwin    # macOS binaries (macOS host only, requires Apple SDK)
make test
```

## Binaries

**`symex`** - CLI symbol dumper.

```sh
$ symex cmd/symex-mcp/glob.go

// cmd/symex-mcp/glob.go
import     path/filepath
import     strings
function   func matchGlob(pattern, path string) bool
           // matchGlob reports whether the slash-separated relative path matches pattern.
           // Supports * (within a segment) and ** (across segments).
```

```sh
$ symex -json cmd/symex-mcp/glob.go

[
  {
    "name": "path/filepath",
    "kind": "import",
    "file": "cmd/symex-mcp/glob.go",
    "line": 4,
    "end_line": 4,
    "signature": "path/filepath"
  },
  {
    "name": "matchGlob",
    "kind": "function",
    "file": "cmd/symex-mcp/glob.go",
    "line": 10,
    "end_line": 49,
    "signature": "func matchGlob(pattern, path string) bool",
    "doc": "// matchGlob reports whether the slash-separated relative path matches pattern.\n// Supports * (within a segment) and ** (across segments)."
  }
]
```

**`symex-mcp`** - MCP server exposing symbol extraction as tools.

```sh
symex-mcp                        # stdio transport (default)
symex-mcp -http                  # HTTP transport on :3002
symex-mcp -http -addr :8080      # custom address
```


## Library

```go
import "indexmcp/symex/extractor"

syms, err := extractor.ExtractPath("/path/to/file.go")
```

`ExtractPath` returns `[]Symbol`, each with `Name`, `Kind`, `File`, `Line`, `EndLine`, `Signature`, and `Doc`.

## System Prompt

Add the following fragment to your LLM system prompt or `CLAUDE.md` to teach the model to use symex-mcp:

```
# Code Navigation

symex-mcp extracts named declarations with signatures, doc strings, and imports.
Prefer these tools over reading source files when exploring a codebase.

If `directory_summary` is available, load all tools first:
ToolSearch `select:directory_summary,file_summary,index,find_symbols,read_symbol,find_importers`

## Workflow

1. `directory_summary(root)` - one line per subdirectory with file + symbol counts; always start here
2. `file_summary(root, glob="subdir/**")` - file-level counts, paged; use `offset`/`limit` (default 50)
3. `index(root, glob, detail="outline"|"full")` - symbol list; outline=name+kind+line, full=+sig+doc
4. `find_symbols(root, pattern="regex", kind?)` - regex on names; `^Foo` prefix, `(?i)foo` case-insensitive
5. `read_symbol(root, name, detail="signature"|"full")` - fetch one symbol's declaration or full body
6. `find_importers(root, pattern="regex")` - which files import a package; `^react`, `/extractor$`

Pass `use_gitignore: true` on all calls in real projects.

## Fall back to Read for
- Non-source files (go.mod, Makefile, config, README)
- Prose or structure between declarations
```

## Adding a Language

Define a `*LangDef` in `extractor/` and register it in `extractor/registry.go`. A `LangDef` specifies the tree-sitter grammar, file extensions, comment node types, and a list of `SymbolDef` / `ImportDef` rules. No imperative handler code is needed for most languages.

See `extractor/go_lang.go` for a representative example.

## Cross-compilation

Linux targets are fully static musl binaries built via `zig cc` and can be cross-compiled from any platform. macOS targets require the Apple SDK and must be built on a macOS host or macOS CI runner.

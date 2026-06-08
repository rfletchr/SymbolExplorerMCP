# symex

Source code symbol extractor and MCP server. Parses source files using tree-sitter and emits named declarations — functions, types, methods, constants — with signatures, doc strings, and import lists. Designed to give LLMs a compact, navigable view of a codebase without reading full file contents.

## Languages

Go, Python, Rust, C, C++, TypeScript, JavaScript

## Building

```sh
make build          # current platform → dist/symex, dist/symex-mcp
make dist-linux     # static Linux binaries via zig cc (any host with zig in PATH)
make dist-darwin    # macOS binaries (macOS host only — requires Apple SDK)
make test
```

## Binaries

**`symex`** — CLI symbol dumper.

```sh
symex ./path/to/project          # text output grouped by file
symex -json ./path/to/project    # JSON array of symbols
```

**`symex-mcp`** — MCP server exposing symbol extraction as tools.

```sh
symex-mcp                        # stdio transport (default)
symex-mcp -http                  # HTTP transport on :3002
symex-mcp -http -addr :8080      # custom address
```

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

## Library

```go
import "indexmcp/symex/extractor"

syms, err := extractor.ExtractPath("/path/to/file.go")
```

`ExtractPath` returns `[]Symbol` — each with `Name`, `Kind`, `File`, `Line`, `EndLine`, `Signature`, and `Doc`.

## Adding a Language

Define a `*LangDef` in `extractor/` and register it in `extractor/registry.go`. A `LangDef` specifies the tree-sitter grammar, file extensions, comment node types, and a list of `SymbolDef` / `ImportDef` rules. No imperative handler code is needed for most languages.

See `extractor/go_lang.go` for a representative example.

## Cross-compilation

Linux targets are fully static musl binaries built via `zig cc` and can be cross-compiled from any platform. macOS targets require the Apple SDK and must be built on a macOS host (or macOS CI runner).

# tiengtrungonline

A command line for [Tieng Trung Online](https://tiengtrungonline.com), a Vietnamese-Chinese learning site.

`tiengtrungonline` is a single pure-Go binary. It reads public data via the WordPress REST API,
shapes it into clean records, and prints output that pipes into the rest of your tools.
No API key, nothing to run alongside it.

## Install

```bash
go install github.com/tamnd/tiengtrungonline-cli/cmd/tiengtrungonline@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/tiengtrungonline-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/tiengtrungonline:latest --help
```

## Usage

```bash
tiengtrungonline posts                        # list recent posts (table)
tiengtrungonline posts -o json                # as JSON, ready for jq
tiengtrungonline posts --per-page 50 --page 2  # second page, 50 per page
tiengtrungonline posts --category 26          # filter by category ID
tiengtrungonline categories                   # list all categories sorted by post count
tiengtrungonline categories -o json           # as JSON
tiengtrungonline --help                       # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

## Commands

### `posts`

List posts from Tieng Trung Online.

Flags:
- `--per-page N` — number of posts per page (default 20)
- `--page N` — page number (default 1)
- `--category ID` — filter by category ID (default 0 = all)

### `categories`

List all categories, sorted by post count descending.

## Development

```
cmd/tiengtrungonline/   thin main
cli/                    command tree (cobra)
tiengtrungonline/       HTTP client and data models
pkg/                    output renderer (table/json/jsonl/csv/tsv/url/raw)
docs/                   tago documentation site
```

```bash
make build      # ./bin/tiengtrungonline
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

## License

Apache-2.0. See [LICENSE](LICENSE).

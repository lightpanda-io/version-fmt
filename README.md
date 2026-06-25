# version-fmt

version-fmt read, format and save versions results.
It creates/updates a version index file with the
[Ziglang](https://ziglang.org/download/index.json) format.
The program is written in [Go](https://go.dev/).

## Usage

The program fetches a GitHub release's assets from `lightpanda-io/browser`,
formats them as a manifest entry and merges it into the index file. The
updated index is written to stdout.

```
version-fmt -release nightly -version 1.2.3-nightly.1456+a1b2c3d index.json
```

### Command line options

- `-release` release tag (default `nightly`)
- `-version` version string from `lightpanda version`

The positional argument is the index file to read and update.

The result is keyed by release tag. Each entry holds the `version`, the build
`date` (taken from the asset `created_at`) and one object per platform
(`download_url`, `shasum`, `size`).

## Docker

Each version of version-fmt is bundled in a docker image available on GH registry.

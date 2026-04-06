# Releasing

Guide for publishing new versions of billmind.

## Requirements

- [goreleaser](https://goreleaser.com/install/) v2+
- `GITHUB_TOKEN` environment variable with `repo` scope
- Git tag following [semver](https://semver.org/) (`v0.1.0`, `v1.0.0`, etc.)

```bash
# Install goreleaser (macOS)
brew install goreleaser

# Verify
goreleaser --version
```

## Release process

### 1. Make sure everything is clean

```bash
go test ./... -race
go build ./...
git status  # should be clean
```

### 2. Choose a version

Follow semver:
- `v0.x.y` — pre-release, API may change
- `vX.0.0` — major (breaking changes)
- `vX.Y.0` — minor (new features, backward compatible)
- `vX.Y.Z` — patch (bug fixes)

### 3. Tag and push

```bash
git tag v0.1.0
git push origin v0.1.0
```

### 4. Run goreleaser

```bash
goreleaser release --clean
```

This will:
- Build binaries for linux/darwin/windows (amd64, arm64, 386)
- Create tar.gz archives (zip for Windows)
- Generate changelog from git commits
- Create a GitHub Release with all artifacts
- Update the Homebrew formula in `curkan/homebrew-public`

### 5. Verify

- Check [GitHub Releases](https://github.com/curkan/billmind/releases)
- Test Homebrew install: `brew install curkan/public/billmind`

## Test build (without publishing)

```bash
goreleaser release --snapshot --clean
```

Builds everything into `dist/` without creating a GitHub release. Useful for verifying the config.

## Troubleshooting

### `multiple tokens found`

```bash
unset GITLAB_TOKEN  # or any other *_TOKEN
goreleaser release --clean
```

### `does not contain a main function`

Make sure `.goreleaser.yaml` has `main: ./cmd/billmind` in the builds section.

### Homebrew formula not updated

- Check that `curkan/homebrew-public` repo exists
- Check that `GITHUB_TOKEN` has write access to that repo

## Artifacts

After a successful release, `dist/` contains:

```
dist/
├── billmind_Darwin_arm64.tar.gz
├── billmind_Darwin_x86_64.tar.gz
├── billmind_Linux_arm64.tar.gz
├── billmind_Linux_x86_64.tar.gz
├── billmind_Linux_i386.tar.gz
├── billmind_Windows_arm64.zip
├── billmind_Windows_x86_64.zip
├── billmind_Windows_i386.zip
├── checksums.txt
└── config.yaml
```

## Installation methods after release

### Homebrew (macOS/Linux)

```bash
brew tap curkan/public
brew install billmind
```

### Manual download

Download the archive for your platform from [Releases](https://github.com/curkan/billmind/releases), extract, and move to PATH:

```bash
tar -xzf billmind_Darwin_arm64.tar.gz
sudo mv billmind /usr/local/bin/
```

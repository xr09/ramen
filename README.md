# Ramen

A simple CLI tool to consume a specified amount of memory. Useful for testing OOM (Out Of Memory) scenarios and resource constraints.

## Installation

### Using Docker
You can run Ramen directly via Docker without installing anything locally. Images are available on GitHub Container Registry (GHCR) and Docker Hub.

```bash
docker run --rm ghcr.io/xr09/ramen -s 1g
```

### From Source
```bash
go install github.com/xr09/ramen/cmd/ramen@latest
```

### Download Binaries
Download the pre-compiled binaries for your platform from the [Releases](https://github.com/xr09/ramen/releases) page.

## Usage

Specify the amount of memory to consume in Megabytes using the `-size` (or also `-s`) flag.

```bash
ramen -size 1024 # Consumes 1GB of RAM
ramen -s 2g      # Consumes 2GB of RAM
```

Press `Ctrl+C` at any time to release the memory and exit.

## Features
- Cross-platform support (Linux, macOS, Windows).
- Efficiently ensures physical memory allocation by touching pages.
- Lightweight with zero external dependencies.

## License
MIT

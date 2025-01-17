# Cluster Codex

An open-source project providing a comprehensive bill of materials (BoM) for Kubernetes clusters, detailing their components, dependencies, and configurations. Simplify cluster management and auditing with a structured and extensible codex.


## Installation

```bash
brew install clx
```
or
```bash
make build
```

### Usage

`clx generate` generates a BOM file for your Kubernetes cluster.

```sh
clx generate [flags]
```

Optional flags include:

```plain
Flags:
  -f, --format string     [optional] cyclonedx-json (default "cyclonedx-json")
  -o, --out-path string   [optional] Path to write generated file to.
  -h, --help              [optional] help for generate
```
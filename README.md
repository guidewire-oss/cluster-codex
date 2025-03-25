# Cluster Codex

An open-source project providing a comprehensive bill of materials (BoM) for Kubernetes clusters, detailing their components, dependencies, and configurations. Simplify cluster management and auditing with a structured and extensible codex.


## Installation

```bash
brew tap guidewire-oss/tap
brew install clx
```
or
```bash
make build
```

## Usage

`clx generate` generates a BOM file for your Kubernetes cluster.

```sh
clx generate [flags]
```

Optional flags include:

```plain
Flags:
  -f, --format string     [optional] cyclonedx-json (default "cyclonedx-json")
  -o, --out-path string   [optional] Path to write generated file to. (default ./output.json)
  -f, --filters string    [optional] Path to a json file containing filters. (default ./filters.json)
  -h, --help              [optional] help for generate
  -l, --log-level string  [optional] Set the logging level (debug, info, warn, error) (default "warn")

```

### Filters
You can specify a file that includes filters. Currently only inclusion filters for namespace are implemented. The 
default filter file is `./filters.json` and `.gitignore` is set to ignore `filter*.json` so that test filters are not
added to the repo by accident. A sample filter file is `./sample-filter.json`.

### Output
Output is written to output.json by default. Here are some useful commands to process that json:
```commandline
# Find all the unique namespaces for components in the output
jq -r '.components[].properties[] | select(.name == "clx:k8s:namespace") | .value' output.json | sort -u

# Find all the applications of kind Namespaces
jq -r '.components[] | select(.type == "application" and any(.properties[]; .name == "clx:k8s:componentKind" and .value == "Namespace")) | .name' output.json | sort
```


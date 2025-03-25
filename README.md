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
  -f, --format string       [optional] cyclonedx-json (default "cyclonedx-json")
  -o, --out-path string     [optional] Path to write generated file to. (default ./output.json)
  -f, --filter-path string  [optional] Path to a json file containing filterPath.
  -h, --help                [optional] help for generate
  -l, --log-level string    [optional] Set the logging level (debug, info, warn, error) (default "warn")

```

### Filters
You can specify a file that includes filterPath. Currently only inclusion filterPath for namespace and kind are implemented. There 
is no default filter file. `.gitignore` is set to ignore `filter*.json` so that if you add a test filter, they are not
added to the repo by accident. 

A sample filter file is `./sample-filter.json`. For the below filter, the BOM will contain `HelmRelease` in all namespaces (`*` in namespaces takes precedence over any other namespace).
```json
{
  "namespaced-inclusions": [
    {
      "namespaces": [
        "test-ns",
        "*"
      ],
      "resources": [
        "HelmRelease"
      ]
    }
  ]
}
```
Similarly for below filter, in addition to `HelmRelease` the BOM will also fetch the `Deployment` across all namespaces (since no namespace is defined for the second filter, it is considered as all namespaces).
```json

{
  "namespaced-inclusions": [
    {
      "namespaces": [
        "test-ns",
        "*"
      ],
      "resources": [
        "HelmRelease"
      ]
    },
    {
      "resources": [ 
        "Deployment"  
      ]
    }
  ]
}
```

For specifying filter for non-namespaced resources like `Namespace`, `PersistentVolume` the below filter can be included in addition to namespaced inclusions.
This will give all the `Namespace` and `PersistentVolume` as well as `Service` and `HelmRelease` from the `kube-system` and `flux-system` namespaces.
```json
{
  "non-namespaced-inclusions": {
    "resources": [
      "Namespace",
      "PersistentVolume"
    ]
  },
  "namespaced-inclusions": [
    {
      "namespaces": [
        "kube-system",
        "flux-system"
      ],
      "resources": [
        "Service",
        "HelmRelease"
      ]
    }
  ]
}
```

### Output
Output is written to output.json by default. Here are some useful commands to process that json:
```commandline
# Find all the unique namespaces for components in the output
jq -r '.components[].properties[] | select(.name == "clx:k8s:namespace") | .value' output.json | sort -u

# Find all the applications of kind Namespaces
jq -r '.components[] | select(.type == "application" and any(.properties[]; .name == "clx:k8s:componentKind" and .value == "Namespace")) | .name' output.json | sort
```


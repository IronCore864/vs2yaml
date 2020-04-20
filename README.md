# vs2yaml

A golang app using vault API to get all secrets under a given path, then use go template to render the secret as a k8s secret yaml files for future uses.

## Build

```
go build
```

## Usage

```
export VAULT_ADDR=https://vault.domain.com
export VAULT_TOKEN=YOUR_TOKEN
export K8S_NAMESPACE=default
# no need to put as "kv/", only "kv" is enough
export VAULT_SECRET_PATH=kv
export OUTPUT_DIR=.
# optional
export VAULT_SKIP_VERIFY=true
./vs2yaml
```

## Docker

```
ironcore864/vs2yaml:latest
```

# vs2yaml

A golang app using vault API to get all secrets under a given path, then use go template to render the secret as a k8s secret yaml files for future uses.

## Build

```
go build
```

## Usage

```
export VAULT_ADDR=https://vault.domain.com
export VAULT_ROLE_ID=ROLE_ID
export VAULT_SECRET_ID=SECRET_ID
export K8S_NAMESPACE=default
# no need to put as "kv/", only "kv" is enough
export VAULT_SECRET_PATH=kv
# kv version, 1 or 2
export VAULT_KV_VERSION=2
export OUTPUT_DIR=.
# optional
export VAULT_SKIP_VERIFY=true
./vs2yaml
```

## Docker

```
ironcore864/vs2yaml:latest
```

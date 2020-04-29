package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"
)

// Config is the struct for environment configuration
type Config struct {
	VaultAddr       string `envconfig:"VAULT_ADDR"`
	K8sNamespace    string `envconfig:"K8S_NAMESPACE"`
	VaultSecretPath string `envconfig:"VAULT_SECRET_PATH"`
	RoleID          string `envconfig:"VAULT_ROLE_ID"`
	SecretID        string `envconfig:"VAULT_SECRET_ID"`
	KvVersion       int    `envconfig:"VAULT_KV_VERSION"`
	OutputDir       string `envconfig:"OUTPUT_DIR"`
}

func main() {
	// load config
	var conf Config
	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	// kv v1 vs v2
	listSecretPath := ""
	readSecretPath := ""
	if conf.KvVersion == 1 {
		listSecretPath = fmt.Sprintf("%s/", conf.VaultSecretPath)
		readSecretPath = fmt.Sprintf("%s", conf.VaultSecretPath)
	} else {
		listSecretPath = fmt.Sprintf("%s/metadata/", conf.VaultSecretPath)
		readSecretPath = fmt.Sprintf("%s/data", conf.VaultSecretPath)
	}

	// vault client
	vaultClient, err := api.NewClient(&api.Config{
		Address: conf.VaultAddr,
	})
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	// AppRole auth
	c := vaultClient.Logical()
	data := map[string]interface{}{
		"role_id":   conf.RoleID,
		"secret_id": conf.SecretID,
	}
	resp, err := c.Write("auth/approle/login", data)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	if resp.Auth == nil {
		log.Fatal("no auth info returned")
		return
	}
	// set token after AppRole auth
	vaultClient.SetToken(resp.Auth.ClientToken)

	// list vault secrets
	secrets, err := c.List(listSecretPath)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	switch x := secrets.Data["keys"].(type) {
	case []interface{}:
		for _, k := range x {
			secretName := fmt.Sprintf("%v", k)
			secret, err := c.Read(fmt.Sprintf("%s/%s", readSecretPath, secretName))
			if err != nil {
				fmt.Println(err)
				return
			}

			data := secret.Data
			for k, v := range data {
				value := fmt.Sprintf("%v", v)
				data[k] = base64.StdEncoding.EncodeToString([]byte(value))
			}

			ctx := map[string]interface{}{
				"name":      secretName,
				"namespace": conf.K8sNamespace,
				"data":      data,
			}

			t, err := template.ParseFiles("secret.yaml.tpl")
			if err != nil {
				log.Fatal(err.Error())
			}
			output, err := os.Create(fmt.Sprintf(fmt.Sprintf("%s/%s.yaml", conf.OutputDir, secretName)))
			if err != nil {
				log.Fatal(err.Error())
			}
			e := t.Execute(output, ctx)
			if e != nil {
				log.Fatal(err.Error())
			}
		}
	}
}

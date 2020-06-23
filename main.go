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

	// kv v2 only
	listSecretPath := fmt.Sprintf("%s/metadata/", conf.VaultSecretPath)
	readSecretPath := fmt.Sprintf("%s/data", conf.VaultSecretPath)

	// vault client
	log.Println("Creating vault client ...")
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
	log.Println("App role auth ...")
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
	log.Println("App role auth succeeded, set token ...")
	vaultClient.SetToken(resp.Auth.ClientToken)

	// list vault secrets
	log.Println("Listing all secrets from vault ...")
	secrets, err := c.List(listSecretPath)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	// iterate all vault secrets, generate k8s secret, and upcert
	log.Println("Starting to create k8s secrets yaml files ...")
	switch x := secrets.Data["keys"].(type) {
	case []interface{}:
		for _, k := range x {
			secretName := fmt.Sprintf("%v", k)
			log.Printf("Processing secret %s ...\n", secretName)
			secret, err := c.Read(fmt.Sprintf("%s/%s", readSecretPath, secretName))
			if err != nil {
				fmt.Println(err)
				return
			}

			data := make(map[string]interface{})
			for k, v := range secret.Data["data"].(map[string]interface{}) {
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
	log.Println("Create/update k8s secrets done!...")
}

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
	Token           string `envconfig:"VAULT_TOKEN"`
	VaultAddr       string `envconfig:"VAULT_ADDR"`
	K8sNamespace    string `envconfig:"K8S_NAMESPACE"`
	VaultSecretPath string `envconfig:"VAULT_SECRET_PATH"`
	OutputDir       string `envconfig:"OUTPUT_DIR"`
}

func main() {
	// load config
	var conf Config
	err := envconfig.Process("vs2yaml", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	// vault client
	client, err := api.NewClient(&api.Config{
		Address: conf.VaultAddr,
	})

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	client.SetToken(conf.Token)

	c := client.Logical()
	secrets, err := c.List(fmt.Sprintf("%s/", conf.VaultSecretPath))
	if err != nil {
		fmt.Println(err)
		return
	}

	switch x := secrets.Data["keys"].(type) {
	case []interface{}:
		for _, k := range x {
			secretName := fmt.Sprintf("%v", k)
			secret, err := c.Read(fmt.Sprintf("%s/%s", conf.VaultSecretPath, secretName))
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

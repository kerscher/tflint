package config

import (
	"os"
	"strings"

	"io/ioutil"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/loader"
)

type Config struct {
	Debug              bool
	DeepCheck          bool              `hcl:"deep_check"`
	AwsCredentials     map[string]string `hcl:"aws_credentials"`
	IgnoreModule       map[string]bool   `hcl:"ignore_module"`
	IgnoreRule         map[string]bool   `hcl:"ignore_rule"`
	Varfile            []string          `hcl:"varfile"`
	TerraformVersion   string            `hcl:"terraform_version"`
	TerraformEnv       string
	TerraformWorkspace string
}

func Init() *Config {
	return &Config{
		Debug:              false,
		DeepCheck:          false,
		AwsCredentials:     map[string]string{},
		IgnoreModule:       map[string]bool{},
		IgnoreRule:         map[string]bool{},
		Varfile:            []string{},
		TerraformEnv:       "default",
		TerraformWorkspace: "default",
	}
}

func (c *Config) LoadConfig(files ...string) error {
	if b, err := ioutil.ReadFile(".terraform/environment"); err == nil {
		c.TerraformEnv = string(b)
		c.TerraformWorkspace = string(b)
	}

	var configs []*ast.ObjectItem
	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			continue
		}

		l := loader.NewLoader(c.Debug)
		if err := l.LoadTemplate(file); err != nil {
			continue
		}

		configs = l.Templates[file].Node.(*ast.ObjectList).Filter("config").Items
		break
	}

	if len(configs) == 0 {
		return nil
	}

	if err := hcl.DecodeObject(c, configs[0]); err != nil {
		return err
	}

	return nil
}

func (c *Config) SetAwsCredentials(accessKey string, secretKey string, profile string, region string) {
	if accessKey != "" {
		c.AwsCredentials["access_key"] = accessKey
	}
	if secretKey != "" {
		c.AwsCredentials["secret_key"] = secretKey
	}
	if profile != "" {
		c.AwsCredentials["profile"] = profile
	}
	if region != "" {
		c.AwsCredentials["region"] = region
	}
}

func (c *Config) HasAwsRegion() bool {
	return c.AwsCredentials["region"] != ""
}

func (c *Config) HasAwsSharedCredentials() bool {
	return c.AwsCredentials["profile"] != "" && c.AwsCredentials["region"] != ""
}

func (c *Config) HasAwsStaticCredentials() bool {
	return c.AwsCredentials["access_key"] != "" && c.AwsCredentials["secret_key"] != "" && c.AwsCredentials["region"] != ""
}

func (c *Config) SetIgnoreModule(ignoreModule string) {
	if ignoreModule == "" {
		return
	}
	ignoreModules := strings.Split(ignoreModule, ",")

	for _, m := range ignoreModules {
		c.IgnoreModule[m] = true
	}
}

func (c *Config) SetIgnoreRule(ignoreRule string) {
	if ignoreRule == "" {
		return
	}
	ignoreRules := strings.Split(ignoreRule, ",")

	for _, r := range ignoreRules {
		c.IgnoreRule[r] = true
	}
}

func (c *Config) SetVarfile(varfile string) {
	// Automatically, `terraform.tfvars` loaded, this priority is the lowest because insert it at the beginning.
	c.Varfile = append([]string{"terraform.tfvars"}, c.Varfile...)

	if varfile == "" {
		return
	}
	varfiles := strings.Split(varfile, ",")
	c.Varfile = append(c.Varfile, varfiles...)
}

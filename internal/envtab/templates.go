package envtab

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type LoadoutTemplate struct {
	Entries     []string `json:"entries" yaml:"entries"`
	Description string   `json:"description" yaml:"description"`
}

type LoadoutTemplates struct {
	Templates map[string]LoadoutTemplate `json:"templates" yaml:"templates"`
}

var envtabTemplates = LoadoutTemplates{
	Templates: map[string]LoadoutTemplate{
		"aws": {
			Description: "Amazon Web Services Template",
			Entries: []string{
				"AWS_ACCESS_KEY_ID",
				"AWS_CA_BUNDLE",
				"AWS_CLI_AUTO_PROMPT",
				"AWS_CLI_FILE_ENCODING",
				"AWS_CONFIG_FILE",
				"AWS_DATA_PATH",
				"AWS_DEFAULT_OUTPUT",
				"AWS_DEFAULT_REGION",
				"AWS_EC2_METADATA_DISABLED",
				"AWS_ENDPOINT_URL",
				"AWS_ENDPOINT_URL_<SERVICE>",
				"AWS_IGNORE_CONFIGURED_ENDPOINT_URLS",
				"AWS_MAX_ATTEMPTS",
				"AWS_METADATA_SERVICE_NUM_ATTEMPTS",
				"AWS_METADATA_SERVICE_TIMEOUT",
				"AWS_PAGER",
				"AWS_PROFILE",
				"AWS_REGION",
				"AWS_RETRY_MODE",
				"AWS_ROLE_ARN",
				"AWS_ROLE_SESSION_NAME",
				"AWS_SECRET_ACCESS_KEY",
				"AWS_SESSION_TOKEN",
				"AWS_SHARED_CREDENTIALS_FILE",
				"AWS_USE_DUALSTACK_ENDPOINT",
				"AWS_USE_FIPS_ENDPOINT",
				"AWS_WEB_IDENTITY_TOKEN_FILE",
			},
		},
	},
}

func MakeLoadoutFromTemplate(templateName string, force bool) Loadout {

	//var templateFound bool

	//for name := range envtabTemplates.Templates {
	//	if name == templateName {
	//		templateFound = true
	//	}
	//}

	// Check for template with short extension
	templatePath := filepath.Join(InitEnvtab(""), "templates/"+templateName+".yml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		slog.Debug("template does not exist", "template", templateName, "path", templatePath)
		os.Exit(1)
	}
	slog.Debug("using template", "template", templateName, "path", templatePath)

	loadout := InitLoadout()

	data, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("ERROR: Failure reading template [%s]: %s\n", templateName, err)
		os.Exit(1)
	}
	// load yaml file into LoadoutTemplate struct
	var template LoadoutTemplate
	err = yaml.Unmarshal(data, &template)
	if err != nil {
		fmt.Printf("ERROR: Failure parsing template [%s]: %s\n", templateName, err)
		os.Exit(1)
	}

	loadout.Metadata.Description = template.Description

	for _, entry := range template.Entries {
		loadout.Entries[entry] = ""
	}

	return *loadout

}

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/templates"
	"github.com/spf13/cobra"
)

var forceFlag bool

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:   "make LOADOUT_NAME TEMPLATE_NAME",
	Short: "Make loadout from a template",
	Long: `Make loadouts from templates. Predefined templates:

Cloud:        aws, gcp, openstack, azure
Databases:    pgsql, mysql, mongodb, elasticsearch
MQ/Msg:       kafka, rabbitmq
Cache:        redis, memcached
Container:    docker, k8s
Secrets:      vault, consul
Tools:        terraform, terragrunt, helm, ansible, packer, vagrant, jira-cli
Languages:    python, go, rust, c
VCS:          git, github, gitlab
Network:      proxy, wireguard
Utils:        sops, yq, jq, jo, etcd, k6

You can also create custom templates in ENVTAB_DIR/templates/ (defaults to $XDG_DATA_HOME/envtab/templates/).`,
	Example:    `  envtab make myloadout aws`,
	Args:       cobra.ExactArgs(2),
	SuggestFor: []string{"create", "new"},
	Aliases:    []string{"m", "mk", "ma", "mak"},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("make called")

		loadoutName := args[0]
		templateName := args[1]

		// Check if loadout already exists
		_, err := backends.ReadLoadout(loadoutName)
		if err == nil {
			// Loadout exists
			if !forceFlag {
				slog.Error("loadout already exists", "loadout", loadoutName)
				os.Exit(1)
			}
		} else if !os.IsNotExist(err) {
			// Some other error occurred
			slog.Error("failure reading loadout", "error", err)
			os.Exit(1)
		}

		loadout := templates.MakeLoadoutFromTemplate(templateName)

		err = backends.WriteLoadout(loadoutName, &loadout)
		if err != nil {
			slog.Error("failure writing loadout", "loadout", loadoutName, "error", err)
			os.Exit(1)
		}

		fmt.Printf("Loadout [%s] created from template [%s]\n", loadoutName, templateName)
		editLoadout(loadoutName)
	},
}

func init() {
	makeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "overwrite existing loadout")
	rootCmd.AddCommand(makeCmd)
}

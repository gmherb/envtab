/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/templates"
	"github.com/spf13/cobra"
)

// mkCmd represents the make command
var mkCmd = &cobra.Command{
	Use:   "make LOADOUT_NAME TEMPLATE_NAME",
	Short: "Make loadout from a template",
	Long: `Make loadouts from templates. Predefined templates:

Cloud:        aws, gcp, azure
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

You can also create custom templates in ~/.envtab/templates/.`,
	Example:    `  envtab make myloadout aws`,
	Args:       cobra.ExactArgs(2),
	SuggestFor: []string{"create", "new"},
	Aliases:    []string{"m", "mk"},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("mk called")

		loadoutName := args[0]
		templateName := args[1]

		loadout := templates.MakeLoadoutFromTemplate(templateName, false)

		err := backends.WriteLoadout(loadoutName, &loadout)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}

		fmt.Printf("Loadout [%s] created from template [%s]\n", loadoutName, templateName)
		editLoadout(loadoutName)
	},
}

func init() {
	rootCmd.AddCommand(mkCmd)
}

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/spf13/cobra"
)

// mkCmd represents the mk command
var mkCmd = &cobra.Command{
	Use:   "mk LOADOUT_NAME TEMPLATE_NAME",
	Short: "Make loadout from a template",
	Long: `Make loadouts from a templates. The following predefined templates
are supported:

Cloud:   Databases:      MQ/Msg:    Cache:      Container:
- aws    - pgsql	     - kafka    - memcached - docker
- gcp    - mysql         - rabbitmq - redis     - k8s
- azure  - mongodb       -          -           -
-        - elasticsearch -          -           -

You can also create your own custom templates and store them in the
templates subdirectory of envtabs HOME.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mk called")

		loadoutName := args[0]
		templateName := args[1]

		loadout := envtab.MakeLoadoutFromTemplate(templateName)

		err := envtab.WriteLoadout(loadoutName, &loadout)
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

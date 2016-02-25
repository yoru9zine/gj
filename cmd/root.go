package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "gj",
	Short: "gj is job manager",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var port int

func init() {
	RootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8181, "port")
}

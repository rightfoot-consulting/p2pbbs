/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "p2pbbs",
	Short: "Provides examples of p2p communications",
	Long: `This application explores the functionality of the libp2p library for golang different commands provide 
different features.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Doesn't do anything at all, it's just there as a placeholder")
}

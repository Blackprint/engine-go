package cmd

import "github.com/spf13/cobra"

var (
	RootCmd = &cobra.Command{
		Use:   "engine-go",
		Short: "engine-go for Blackprint",
		Long:  "engine-go is Blackprint Engine for Golang",
	}
)

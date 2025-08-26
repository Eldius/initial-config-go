package setup

import (
	"github.com/spf13/cobra"
)

func PersistentPreRunE(appName string, opts ...OptionFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return InitSetup(appName, opts...)
	}
}

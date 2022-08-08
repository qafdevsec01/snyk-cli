package cliv2

import (
	"fmt"

	"github.com/snyk/cli-extension-lib-go/extension"
	"github.com/snyk/cli/cliv2/internal/httpauth"
	"github.com/spf13/cobra"
)

const (
	CMDARG_PROXY_NO_AUTH string = "proxy-noauth"
)

func MakeArgParserConfig(extensions []*extension.Extension, config *CliConfiguration) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "snyk",
		Short: "Snyk CLI scans and monitors your projects for security vulnerabilities and license issues.",
		Long:  `Snyk CLI scans and monitors your projects for security vulnerabilities and license issues.`,
	}

	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "d", false, "Enable debug logging.")
	rootCmd.PersistentFlags().BoolVar(&config.Insecure, "insecure", false, "Disable secure communication protocols.")
	rootCmd.PersistentFlags().Bool(CMDARG_PROXY_NO_AUTH, false, "Disable all proxy authentication.")
	rootCmd.PersistentFlags().StringVar(&config.ProxyAddr, "proxy", "", "Configure a http/https proxy. Overriding environment variables.")

	// add a command for each of the extensions
	for _, x := range extensions {
		config.DebugLogger.Println("adding extension to arg parser:", x.Metadata.Name)
		c := cobraCommandFromExtensionMetadataCommand(x.Metadata.Command)
		rootCmd.AddCommand(c)
	}

	// add top-level commands for CLIv1
	addV1TopLevelCommands(rootCmd)

	return rootCmd
}

var v1Commands = []string{
	"test",
	"monitor",
	"code",
	"iac",
	"container",
	"auth",
	// TODO: what else?
}

func addV1TopLevelCommands(rootCmd *cobra.Command) {
	for _, command := range v1Commands {
		c := &cobra.Command{
			Use:   command,
			Short: "todo: command description",
			Long:  "todo: command description",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("v1 command %s called\n", command)
				fmt.Println("args:", args)
			},
		}

		// c.SilenceUsage = true
		// c.SilenceErrors = true

		// because we don't know anything about the command internals - they are all handled by v1
		c.DisableFlagParsing = true

		rootCmd.AddCommand(c)
	}
}

func cobraCommandFromExtensionMetadataCommand(cmd *extension.Command) *cobra.Command {
	cobraCommand := &cobra.Command{
		Use:   cmd.Name,
		Short: cmd.Description,
		Long:  cmd.Description,
		// this `Run` field is required in order to make this command show in the "available commands" list in the usage
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Add options
	if cmd.Options != nil {
		for _, o := range cmd.Options {
			if o.Type == "string" {
				cobraCommand.Flags().StringP(o.Name, o.Shorthand, "", o.Description)
			} else if o.Type == "bool" {
				cobraCommand.Flags().BoolP(o.Name, o.Shorthand, false, o.Description)
			}
		}
	}

	// Add debug option
	cobraCommand.Flags().BoolP("debug", "d", false, "debug mode")

	// Add subcommands
	if cmd.Subcommands != nil {
		for _, sc := range cmd.Subcommands {
			cobraSubcommand := cobraCommandFromExtensionMetadataCommand(sc)
			cobraCommand.AddCommand(cobraSubcommand)
		}
	}

	return cobraCommand
}

func ExecuteArgumentParser(argParserRootCmd *cobra.Command, config *CliConfiguration) (err error) {
	err = argParserRootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if isSet, _ := argParserRootCmd.PersistentFlags().GetBool(CMDARG_PROXY_NO_AUTH); isSet {
		config.ProxyAuthenticationMechanism = httpauth.NoAuth
	}
	return err
}

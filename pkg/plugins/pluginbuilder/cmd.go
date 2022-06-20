package pluginbuilder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter/version"
	"github.com/spf13/cobra"
)

func main() {
	opts := PluginOptions{
		Name:              "myplugin",
		RegisteredPlugins: nil,
		Version:           "v1",
		Commit:            "abc123",
	}
	ctx := context.Background()
	app := NewPlugin(opts)
	rootCmd := buildPluginCommand(app)
	cli.Main(ctx, rootCmd, app)
}

func buildPluginCommand(p *PorterPlugin) *cobra.Command {
	p.porterConfig.In = getInput()

	cmd := &cobra.Command{
		Use:   p.Name(),
		Short: fmt.Sprintf("%s plugin for porter", p.Name()),
	}

	cmd.AddCommand(buildVersionCommand(p))
	cmd.AddCommand(buildRunCommand(p))

	return cmd
}

func buildRunCommand(p *PorterPlugin) *cobra.Command {
	opts := RunOptions{}
	cmd := &cobra.Command{
		Use:   "run PLUGIN_KEY",
		Short: "Run the plugin and listen for client connections",
		Long: `Run the specified PLUGIN_KEY and listen for client connections.

PLUGIN_KEY should be the fully-qualified 3-part key for the requested plugin which follows the format: "INTERFACE.NAME.IMPLEMENTATION".
For example, "storage.porter.mongodb" or "secrets.hashicorp.vault". 
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.ApplyArgs(args); err != nil {
				return err
			}
			return p.Run(cmd.Context(), opts)
		},
	}

	return cmd
}

func buildVersionCommand(p *PorterPlugin) *cobra.Command {
	opts := version.Options{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the plugin version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintVersion(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", string(version.DefaultVersionFormat),
		"Specify an output format.  Allowed values: json, plaintext")

	return cmd
}

// getInput attempts to use os.Stdin for standard input
// otherwise it returns an empty buffer
func getInput() io.Reader {
	s, _ := os.Stdin.Stat()
	if (s.Mode() & os.ModeCharDevice) == 0 {
		return os.Stdin
	}

	return &bytes.Buffer{}
}

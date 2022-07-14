package main

import (
	"strings"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildAliasCommands(p *porter.Porter) []*cobra.Command {
	return []*cobra.Command{
		buildCreateAlias(p),
		buildBuildAlias(p),
		buildLintAlias(p),
		buildInstallAlias(p),
		buildUpgradeAlias(p),
		buildUninstallAlias(p),
		buildInvokeAlias(p),
		buildPublishAlias(p),
		buildListAlias(p),
		buildShowAlias(p),
		buildArchiveAlias(p),
		buildExplainAlias(p),
		buildCopyAlias(p),
		buildInspectAlias(p),
		buildLogsAlias(p),
	}
}

func buildCreateAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleCreateCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle create", "porter create", -1)
	cli.SetCommandGroup(cmd, "alias")
	return cmd
}

func buildBuildAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleBuildCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle build", "porter build", -1)
	cli.SetCommandGroup(cmd, "alias")
	return cmd
}

func buildLintAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleLintCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle lint", "porter lint", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildInstallAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleInstallCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle install", "porter install", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildUpgradeAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleUpgradeCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle upgrade", "porter upgrade", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildInvokeAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleInvokeCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle invoke", "porter invoke", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildUninstallAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleUninstallCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle uninstall", "porter uninstall", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildPublishAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundlePublishCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle publish", "porter publish", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildShowAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationShowCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter installation show", "porter show", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildListAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationsListCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter installations list", "porter list", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildArchiveAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleArchiveCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle archive", "porter archive", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildExplainAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleExplainCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle explain", "porter explain", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildCopyAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleCopyCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle copy", "porter copy", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildInspectAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleInspectCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle inspect", "porter inspect", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

func buildLogsAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationLogShowCommand(p)
	cmd.Use = "logs"
	cmd.Aliases = []string{"log"}
	cmd.Example = strings.Replace(cmd.Example, "porter installation logs show", "porter logs", -1)
	cli.SetCommandGroup(cmd, "alias")

	return cmd
}

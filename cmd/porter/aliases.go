package main

import (
	"strings"

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
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle create", "porter create")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildBuildAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleBuildCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle build", "porter build")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildLintAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleLintCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle lint", "porter lint")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildInstallAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationInstallCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installation install", "porter install")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildUpgradeAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationUpgradeCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installation upgrade", "porter upgrade")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildInvokeAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationInvokeCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installation invoke", "porter invoke")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildUninstallAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationUninstallCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installation uninstall", "porter uninstall")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildPublishAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundlePublishCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle publish", "porter publish")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildShowAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationShowCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installation show", "porter show")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildListAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationsListCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installations list", "porter list")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildArchiveAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleArchiveCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle archive", "porter archive")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildExplainAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleExplainCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle explain", "porter explain")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildCopyAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleCopyCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle copy", "porter copy")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildInspectAlias(p *porter.Porter) *cobra.Command {
	cmd := buildBundleInspectCommand(p)
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter bundle inspect", "porter inspect")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildLogsAlias(p *porter.Porter) *cobra.Command {
	cmd := buildInstallationLogShowCommand(p)
	cmd.Use = "logs"
	cmd.Aliases = []string{"log"}
	cmd.Example = strings.ReplaceAll(cmd.Example, "porter installation logs show", "porter logs")
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

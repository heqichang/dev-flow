package cmd

import (
	"os"

	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "生成 shell 补全脚本",
		Long: `生成指定 shell 的自动补全脚本。

安装方法：

Bash:
  $ devflow completion bash > /etc/bash_completion.d/devflow
  或
  $ devflow completion bash > ~/.bash_completion.d/devflow

Zsh:
  $ devflow completion zsh > ~/.zsh/completion/_devflow

Fish:
  $ devflow completion fish > ~/.config/fish/completions/devflow.fish

PowerShell:
  PS> devflow completion powershell >> $PROFILE`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				ui.Error("不支持的 shell 类型：" + args[0])
				os.Exit(1)
			}
		},
	}
	return cmd
}

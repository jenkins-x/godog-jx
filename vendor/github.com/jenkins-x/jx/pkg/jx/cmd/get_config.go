package cmd

import (
	"io"

	"github.com/jenkins-x/jx/pkg/config"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
)

// GetConfigOptions the command line options
type GetConfigOptions struct {
	GetOptions

	Dir string
}

var (
	getConfigLong = templates.LongDesc(`
		Display the project configuration

`)

	getConfigExample = templates.Examples(`
		# View the project configuration
		jx get config
	`)
)

// NewCmdGetConfig creates the command
func NewCmdGetConfig(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &GetConfigOptions{
		GetOptions: GetOptions{
			CommonOptions: CommonOptions{
				Factory: f,
				Out:     out,
				Err:     errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:     "config [flags]",
		Short:   "Display the project configuration",
		Long:    getConfigLong,
		Example: getConfigExample,
		Aliases: []string{"url"},
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}
	options.addGetConfigFlags(cmd)
	return cmd
}

func (o *GetConfigOptions) addGetConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Dir, "dir", "d", "", "The root project directory")
}

// Run implements this command
func (o *GetConfigOptions) Run() error {
	pc, _, err := config.LoadProjectConfig(o.Dir)
	if err != nil {
		return err
	}
	if pc.IsEmpty() {
		o.Printf("No project configuration for this directory.\n")
		o.Printf("To edit the configuration use: %s\n", util.ColorInfo("jx edit config"))
		return nil
	}
	table := o.CreateTable()
	table.AddRow("Service", "Kind", "URL", "NAME")

	t := pc.IssueTracker
	if t != nil {
		table.AddRow("Issue Tracker", t.Kind, t.URL, t.Project)
	}
	w := pc.Wiki
	if w != nil {
		table.AddRow("Wiki", w.Kind, w.URL, w.Space)
	}
	ch := pc.Chat
	if w != nil {
		table.AddRow("Chat", ch.Kind, ch.URL, ch.Room)
	}
	table.Render()
	return nil
}

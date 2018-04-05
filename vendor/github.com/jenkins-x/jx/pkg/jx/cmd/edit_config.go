package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx/pkg/config"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
)

const (
	chatKind  = "chat"
	issueKind = "issues"
	wikiKind  = "wiki"
)

var (
	editConfigLong = templates.LongDesc(`
		Edits the project configuration
`)

	editConfigExample = templates.Examples(`
		# Edit the project configuration for the current directory
		jx edit config
	`)

	configKinds = []string{
		chatKind,
		issueKind,
		wikiKind,
	}
)

// EditConfigOptions the options for the create spring command
type EditConfigOptions struct {
	EditOptions

	Dir  string
	Kind string

	IssuesAuthConfigSvc auth.AuthConfigService
}

// NewCmdEditConfig creates a command object for the "create" command
func NewCmdEditConfig(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &EditConfigOptions{
		EditOptions: EditOptions{
			CommonOptions: CommonOptions{
				Factory: f,
				Out:     out,
				Err:     errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Edits the project configuration",
		Aliases: []string{"project"},
		Long:    editConfigLong,
		Example: editConfigExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&options.Dir, "dir", "d", "", "The root project directory")
	cmd.Flags().StringVarP(&options.Kind, "kind", "k", "", "The kind of configuration to edit root project directory. Possible values "+strings.Join(configKinds, ", "))

	return cmd
}

// Run implements the command
func (o *EditConfigOptions) Run() error {
	authConfigSvc, err := o.Factory.CreateIssueTrackerAuthConfigService()
	if err != nil {
		return err
	}
	o.IssuesAuthConfigSvc = authConfigSvc

	pc, fileName, err := config.LoadProjectConfig(o.Dir)
	if err != nil {
		return err
	}

	kind := o.Kind
	if kind == "" && !o.BatchMode {
		kind, err = util.PickRequiredNameWithDefault(configKinds, "Which configuration do you want to edit", issueKind)
		if err != nil {
			return err
		}
	}
	if kind == "" {
		return fmt.Errorf("No kind option!")
	}
	if util.StringArrayIndex(configKinds, kind) < 0 {
		return util.InvalidOption("kind", kind, configKinds)
	}
	modified := false
	switch kind {
	case issueKind:
		modified, err = o.EditIssueTracker(pc)
	default:
		return fmt.Errorf("Editing %s is not yet supported!", kind)
	}
	if modified {
		err = pc.SaveConfig(fileName)
		if err != nil {
			return err
		}
		o.Printf("Saved project configuration %s\n", util.ColorInfo(fileName))
	}
	return nil
}

func (o *EditConfigOptions) EditIssueTracker(pc *config.ProjectConfig) (bool, error) {
	answer := false
	if pc.IssueTracker == nil {
		pc.IssueTracker = &config.IssueTrackerConfig{}
		answer = true
	}
	it := pc.IssueTracker

	config := o.IssuesAuthConfigSvc.Config()
	if len(config.Servers) == 0 {
		return answer, fmt.Errorf("No issue tracker servers available. Please add one via: jx create tracker server")
	}
	server, err := config.PickServer("Issue tracker service", o.BatchMode)
	if err != nil {
		return answer, err
	}
	if server == nil || server.URL == "" {
		return answer, fmt.Errorf("No issue tracker server URL found!")
	}
	it.URL = server.URL
	if server.Kind != "" {
		it.Kind = server.Kind
	}
	answer = true

	it.Project, err = util.PickValue("Issue tracker project name: ", it.Project, true)
	if err != nil {
		return answer, err
	}
	return answer, nil
}

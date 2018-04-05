package cmd

import (
	"fmt"
	"io"

	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
)

var (
	createTrackerServer_long = templates.LongDesc(`
		Adds a new Issue Tracker Server URL
`)

	createTrackerServer_example = templates.Examples(`
		# Add a new git server URL
		jx create tracker server gitea
	`)

	trackerKindToServiceName = map[string]string{
		"bitbucket": "bitbucket-bitbucket",
	}
)

// CreateTrackerServerOptions the options for the create spring command
type CreateTrackerServerOptions struct {
	CreateOptions

	Name string
}

// NewCmdCreateTrackerServer creates a command object for the "create" command
func NewCmdCreateTrackerServer(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &CreateTrackerServerOptions{
		CreateOptions: CreateOptions{
			CommonOptions: CommonOptions{
				Factory: f,
				Out:     out,
				Err:     errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:     "server kind [url]",
		Short:   "Creates a new issue tracker server URL",
		Aliases: []string{"provider"},
		Long:    createTrackerServer_long,
		Example: createTrackerServer_example,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.Name, "name", "n", "", "The name for the issue tracker server being created")
	return cmd
}

// Run implements the command
func (o *CreateTrackerServerOptions) Run() error {
	args := o.Args
	if len(args) < 1 {
		return missingTrackerArguments()
	}
	kind := args[0]
	name := o.Name
	if name == "" {
		name = kind
	}
	gitUrl := ""
	if len(args) > 1 {
		gitUrl = args[1]
	} else {
		// lets try find the git URL based on the provider
		serviceName := trackerKindToServiceName[kind]
		if serviceName != "" {
			url, err := o.findService(serviceName)
			if err != nil {
				return fmt.Errorf("Failed to find %s issue tracker serivce %s: %s", kind, serviceName, err)
			}
			gitUrl = url
		}
	}

	if gitUrl == "" {
		return missingTrackerArguments()
	}
	authConfigSvc, err := o.Factory.CreateIssueTrackerAuthConfigService()
	if err != nil {
		return err
	}
	config := authConfigSvc.Config()
	config.GetOrCreateServerName(gitUrl, name, kind)
	config.CurrentServer = gitUrl
	err = authConfigSvc.SaveConfig()
	if err != nil {
		return err
	}
	o.Printf("Added issue tracker server %s for URL %s\n", util.ColorInfo(name), util.ColorInfo(gitUrl))
	return nil
}

func missingTrackerArguments() error {
	return fmt.Errorf("Missing tracker server URL arguments. Usage: jx create tracker server kind [url]")
}

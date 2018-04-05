package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/kube"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/pkg/browser"
)

type ConsoleOptions struct {
	GetURLOptions

	OnlyViewURL bool
}

var (
	console_long = templates.LongDesc(`
		Opens the Jenkins X console in a browser.`)
	console_example = templates.Examples(`
		# Open the Jenkins X console in a browser
		jx console

		# Print the Jenkins X console URL but do not open a browser
		jx console -u`)
)

func NewCmdConsole(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &ConsoleOptions{
		GetURLOptions: GetURLOptions{
			GetOptions: GetOptions{
				CommonOptions: CommonOptions{
					Factory: f,
					Out:     out,
					Err:     errOut,
				},
			},
		},
	}
	cmd := &cobra.Command{
		Use:     "console",
		Short:   "Opens the Jenkins console",
		Long:    console_long,
		Example: console_example,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}
	options.addConsoleFlags(cmd)
	return cmd
}

func (o *ConsoleOptions) addConsoleFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.OnlyViewURL, "url", "u", false, "Only displays and the URL and does not open the browser")

	o.addGetUrlFlags(cmd)
}

func (o *ConsoleOptions) Run() error {
	return o.Open(kube.ServiceJenkins, "Jenkins Console")
}

func (o *ConsoleOptions) Open(name string, label string) error {
	var err error
	url := ""
	ns := o.Namespace
	if ns == "" && o.Environment != "" {
		ns, err = o.findEnvironmentNamespace(o.Environment)
		if err != nil {
			return err
		}
	}
	if ns != "" {
		url, err = o.findServiceInNamespace(name, ns)
	} else {
		url, err = o.findService(name)
	}
	if err != nil && name != "" {
		o.Printf("If the app %s is running in a different environment you could try: %s\n", util.ColorInfo(name), util.ColorInfo("jx get applications"))
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(o.Out, "%s: %s\n", label, util.ColorInfo(url))
	if !o.OnlyViewURL {
		browser.OpenURL(url)
	}
	return nil
}

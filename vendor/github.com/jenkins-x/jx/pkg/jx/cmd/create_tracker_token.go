package cmd

import (
	"fmt"
	"io"

	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/issues"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/kube"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	createTrackerTokenLong = templates.LongDesc(`
		Creates a new User Token for an Issue Tracker
`)

	createTrackerTokenExample = templates.Examples(`
		# Add a new User Token for an Issue Tracker
		jx create tracker token -n jira someUserName

		# As above with the password being passed in
		jx create git token -n jira -p somePassword someUserName	
	`)
)

// CreateTrackerTokenOptions the command line options for the command
type CreateTrackerTokenOptions struct {
	CreateOptions

	ServerFlags ServerFlags
	Username    string
	Password    string
	ApiToken    string
	Timeout     string
}

// NewCmdCreateTrackerToken creates a command
func NewCmdCreateTrackerToken(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &CreateTrackerTokenOptions{
		CreateOptions: CreateOptions{
			CommonOptions: CommonOptions{
				Factory: f,
				Out:     out,
				Err:     errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:     "token [username]",
		Short:   "Adds a new token/login for a user on an issue tracker server",
		Aliases: []string{"login"},
		Long:    createTrackerTokenLong,
		Example: createTrackerTokenExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}
	options.addCommonFlags(cmd)
	options.ServerFlags.addGitServerFlags(cmd)
	cmd.Flags().StringVarP(&options.ApiToken, "api-token", "t", "", "The API Token for the user")
	cmd.Flags().StringVarP(&options.Timeout, "timeout", "", "", "The timeout if using browser automation to generate the API token (by passing username and password)")

	return cmd
}

// Run implements the command
func (o *CreateTrackerTokenOptions) Run() error {
	args := o.Args
	if len(args) > 0 {
		o.Username = args[0]
	}
	if len(args) > 1 {
		o.ApiToken = args[1]
	}
	authConfigSvc, err := o.Factory.CreateIssueTrackerAuthConfigService()
	if err != nil {
		return err
	}
	config := authConfigSvc.Config()

	server, err := o.findIssueTrackerServer(config, &o.ServerFlags)
	if err != nil {
		return err
	}
	if o.Username == "" {
		return fmt.Errorf("No Username specified")
	}

	userAuth := config.GetOrCreateUserAuth(server.URL, o.Username)
	if o.ApiToken != "" {
		userAuth.ApiToken = o.ApiToken
	}

	tokenUrl := issues.ProviderAccessTokenURL(server.Kind, server.URL)

	if userAuth.IsInvalid() {

		f := func(username string) error {
			o.Printf("Please generate an API Token for server %s\n", server.Label())
			if tokenUrl != "" {
				o.Printf("Click this URL %s\n\n", util.ColorInfo(tokenUrl))
			}
			o.Printf("Then COPY the token and enter in into the form below:\n\n")
			return nil
		}

		err = config.EditUserAuth(server.Label(), userAuth, o.Username, false, o.BatchMode, f)
		if err != nil {
			return err
		}
		if userAuth.IsInvalid() {
			return fmt.Errorf("You did not properly define the user authentication!")
		}
	}

	config.CurrentServer = server.URL
	err = authConfigSvc.SaveConfig()
	if err != nil {
		return err
	}

	err = o.updateIssueTrackerCredentialsSecret(server, userAuth)
	if err != nil {
		o.warnf("Failed to update jenkins issue tracker credentials secret: %v\n", err)
	}

	o.Printf("Created user %s API Token for git server %s at %s\n",
		util.ColorInfo(o.Username), util.ColorInfo(server.Name), util.ColorInfo(server.URL))
	return nil
}

func (o *CreateTrackerTokenOptions) updateIssueTrackerCredentialsSecret(server *auth.AuthServer, userAuth *auth.UserAuth) error {
	client, curNs, err := o.Factory.CreateClient()
	if err != nil {
		return err
	}
	ns, _, err := kube.GetDevNamespace(client, curNs)
	if err != nil {
		return err
	}

	options := metav1.GetOptions{}
	name := kube.SecretJenkinsIssueCredentials + server.Kind
	secrets := client.CoreV1().Secrets(ns)
	secret, err := secrets.Get(name, options)
	create := false
	operation := "update"
	if err != nil {
		// lets create a new secret
		create = true
		operation = "create"
		secret = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Data: map[string][]byte{},
		}
	}
	if userAuth.Username != "" {
		secret.Data["username"] = []byte(userAuth.Username)
	}
	if userAuth.ApiToken != "" {
		secret.Data["token"] = []byte(userAuth.ApiToken)
	}
	if create {
		_, err = secrets.Create(secret)
	} else {
		_, err = secrets.Update(secret)
	}
	if err != nil {
		return fmt.Errorf("Failed to %s secret %s due to %s", operation, secret.Name, err)
	}
	return nil
}

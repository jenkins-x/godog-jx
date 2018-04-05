package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jenkins-x/godog-jx/jenkins"
	"github.com/jenkins-x/godog-jx/utils"
	"github.com/jenkins-x/golang-jenkins"
	"github.com/jenkins-x/jx/pkg/gits"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
)

type CommonTest struct {
	Factory       cmdutil.Factory
	JenkinsClient *gojenkins.Jenkins
	Interactive   bool
	Errors        *utils.ErrorSlice
	WorkDir       string
	AppName       string
	Organisation  string
}

// TheApplicationShouldBeBuiltAndPromotedViaCICD asserts that the project
// should be created in Jenkins and that the build should complete successfully
func (o *CommonTest) TheApplicationShouldBeBuiltAndPromotedViaCICD() error {
	appName := o.AppName
	if appName == "" {
		_, appName = filepath.Split(o.WorkDir)
	}
	f := o.Factory
	gitURL, err := o.GitProviderURL()
	if err != nil {
		return err
	}
	gitAuthSvc, err := f.CreateGitAuthConfigService()
	if err != nil {
		return err
	}
	gitConfig := gitAuthSvc.Config()
	server := gitConfig.GetServer(gitURL)
	if server == nil {
		return fmt.Errorf("Could not find a git auth user for git server URL %s", gitURL)
	}
	owner := o.GetGitOrganisation()
	if owner == "" {
		userName := server.CurrentUser
		if userName == "" {
			if len(server.Users) == 0 {
				return fmt.Errorf("No users are configured for authentication with git server URL %s", gitURL)
			}
			userName = server.Users[0].Username
		}
		if userName == "" {
			return fmt.Errorf("Could not detect username for git server URL %s", gitURL)
		}
		owner = userName
	}
	jobName := owner + "/" + appName + "/master"
	if o.JenkinsClient == nil {
		client, err := f.CreateJenkinsClient()
		if err != nil {
			return err
		}
		o.JenkinsClient = client
	}
	fmt.Printf("Checking that there is a job built successfully for %s\n", jobName)
	return jenkins.ThereShouldBeAJobThatCompletesSuccessfully(jobName, o.JenkinsClient)
}

// TheGitInfoShouldHaveAJobAndShouldBeBuiltAndPromotedViaCICD asserts that the project
// should be created in Jenkins and that the build should complete successfully
func (o *CommonTest) TheGitInfoShouldHaveAJobAndShouldBeBuiltAndPromotedViaCICD() error {
	url, err := gits.DiscoverRemoteGitURL(filepath.Join(o.WorkDir, ".git/config"))
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("Could not discover the remote git URL")
	}
	gitInfo, err := gits.ParseGitURL(url)
	if err != nil {
		return err
	}

	owner := gitInfo.Organisation
	repo := gitInfo.Name
	if owner == "" {
		return fmt.Errorf("Could not find owner from GitInfo %#v from URL %s", gitInfo, url)
	}
	if repo == "" {
		return fmt.Errorf("Could not find repo name from GitInfo %#v from URL %s", gitInfo, url)
	}
	jobName := owner + "/" + repo + "/master"
	if o.JenkinsClient == nil {
		client, err := o.Factory.CreateJenkinsClient()
		if err != nil {
			return err
		}
		o.JenkinsClient = client
	}
	fmt.Printf("Checking that there is a job built successfully for %s\n", jobName)
	return jenkins.ThereShouldBeAJobThatCompletesSuccessfully(jobName, o.JenkinsClient)
}

// GitProviderURLOrEmpty returns the git provider to use based on the environment variable
// `GIT_PROVIDER_URL` or returns the empty string
func (o *CommonTest) GitProviderURLOrEmpty() (string, error) {
	gitProviderURL := os.Getenv("GIT_PROVIDER_URL")
	if gitProviderURL != "" {
		return gitProviderURL, nil
	}
	return "", nil
}

// GitProviderURL returns the git provider URL to use based on the environment variable
// `GIT_PROVIDER_URL` or returns the default server's URL
func (o *CommonTest) GitProviderURL() (string, error) {
	gitProviderURL := os.Getenv("GIT_PROVIDER_URL")
	if gitProviderURL != "" {
		return gitProviderURL, nil
	}
	// find the default load the default one from the current ~/.jx/jenkinsAuth.yaml
	authConfigSvc, err := o.Factory.CreateGitAuthConfigService()
	if err != nil {
		return "", err
	}
	config := authConfigSvc.Config()
	url := config.CurrentServer
	if url != "" {
		return url, nil
	}
	servers := config.Servers
	if len(servers) == 0 {
		return "", fmt.Errorf("No servers in the ~/.jx/gitAuth.yaml file!")
	}
	return servers[0].URL, nil
}

// GetGitOrganisation returns the git organisation to create new projects inside
func (o* CommonTest) GetGitOrganisation() string {
	org := o.Organisation
	if org == "" {
		org = os.Getenv("GIT_ORGANISATION")
	}
	if org == "" {
		org = "jenkins-x-tests"
	}
	return org
}
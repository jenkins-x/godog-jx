package godog_jx

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/common"
	"github.com/jenkins-x/godog-jx/utils"
	"github.com/jenkins-x/jx/pkg/gits"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
)

type importurlTest struct {
	common.CommonTest

	Args []string
}

func (o *importurlTest) aTemporaryWorkingDirectory() error {
	tempDirPrefix := "import-url-"
	tmpDir, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return err
	}
	err = os.MkdirAll(tmpDir, utils.DefaultWritePermissions)
	if err != nil {
		return err
	}
	o.WorkDir = tmpDir
	return o.Errors.Error()
}

func (o *importurlTest) runningJxImportUrlWithInADirectory(url string) error {
	assert.NotEmpty(o.Errors, url, "no URL provided!")
	gitInfo, err := gits.ParseGitURL(url)
	if err != nil {
		return fmt.Errorf("Failed to parse git URL %s: %s", url, err)
	}

	// lets try find if there's a Jenkins project for the current repo
	owner := gitInfo.Organisation
	repo := gitInfo.Name
	if owner == "" {
		return fmt.Errorf("Could not find owner from GitInfo %#v from URL %s", gitInfo, url)
	}
	if repo == "" {
		return fmt.Errorf("Could not find repo name from GitInfo %#v from URL %s", gitInfo, url)
	}
	jobName := owner + "/" + repo
	if o.JenkinsClient == nil {
		client, err := o.Factory.CreateJenkinsClient()
		if err != nil {
			return err
		}
		o.JenkinsClient = client
	}
	job, err := o.JenkinsClient.GetJob(jobName)
	if err != nil {
		fmt.Printf("Could not find a job called %s: %s", jobName, err)
	} else {
		err := o.JenkinsClient.DeleteJob(job)
		if err != nil {
			return fmt.Errorf("Failed to delete Jenkins job %s: %s", job.Url, err)
		}
		fmt.Printf("Deleted jenkins job %s\n", job.Url)
	}

	remaining := []string{"import", "--url", url, "-b"}

	// allow custom git providers
	gitProviderURL, err := o.GitProviderURLOrEmpty()
	if err != nil {
		return err
	}
	if gitProviderURL != "" {
		fmt.Printf("Using git provider URL %s\n", util.ColorInfo(gitProviderURL))
		remaining = append(remaining, "--git-provider-url", gitProviderURL)
	}
	if len(o.Args) > 0 {
		remaining = append(remaining, o.Args...)
	}
	fmt.Printf("Running jx %s\n", strings.Join(remaining, " "))
	err = utils.RunCommandInteractive(o.Interactive, o.WorkDir, "jx", remaining...)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func (o *importurlTest) thereShouldBeAJenkinsProjectCreate() error {
	// lets update the work directory to the newly cloned repo
	files, err := ioutil.ReadDir(o.WorkDir)
	if err != nil {
		return fmt.Errorf("FAiled to read directory %s: %s", o.WorkDir, err)
	}
	for _, f := range files {
		if f.IsDir() {
			o.WorkDir = filepath.Join(o.WorkDir, f.Name())
			fmt.Printf("Setting work directory to %s\n", o.WorkDir)
			return nil
		}
	}
	return fmt.Errorf("Could not find a child work directory inside %s", o.WorkDir)
}

func ImporturlFeatureContext(s *godog.Suite) {
	o := &importurlTest{
		CommonTest: common.CommonTest{
			Factory:     cmdutil.NewFactory(),
			Interactive: os.Getenv("JX_INTERACTIVE") == "true",
			Errors:      utils.CreateErrorSlice(),
		},
		Args: []string{},
	}
	s.Step(`^a temporary directory$`, o.aTemporaryWorkingDirectory)
	s.Step(`^running \'jx import --url\' with "([^"]*)" in a directory$`, o.runningJxImportUrlWithInADirectory)
	s.Step(`^there should be a jenkins project created$`, o.thereShouldBeAJenkinsProjectCreate)
	s.Step(`^the application should be built and promoted via CI \/ CD$`, o.TheGitInfoShouldHaveAJobAndShouldBeBuiltAndPromotedViaCICD)
}

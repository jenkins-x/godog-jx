package godog_jx

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/common"
	"github.com/jenkins-x/godog-jx/utils"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
)

type golangHTTPTest struct {
	common.Test

	Args []string
}

func (o *golangHTTPTest) aWorkDirectory() error {
	var err error
	tmpDir, err = ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return err
	}
	err = os.MkdirAll(tmpDir, utils.DefaultWritePermissions)
	if err != nil {
		return err
	}
	o.WorkDir = tmpDir
	assert.DirExists(o.Errors, o.WorkDir)
	return o.Errors.Error()
}

func (o *golangHTTPTest) runningInThatDirectory(commandLine string) error {
	args := strings.Fields(commandLine)
	assert.NotEmpty(o.Errors, args, "not enough arguments")
	cmd := args[0]
	assert.Equal(o.Errors, "jx", cmd)
	gitProviderURL, err := o.GitProviderURL()
	if err != nil {
		return err
	}
	fmt.Printf("Using git provider URL %s and work directory %s\n", util.ColorInfo(gitProviderURL), util.ColorInfo(o.WorkDir))
	remaining := append(args[1:], "--org", o.GetGitOrganisation())
	if len(o.Args) > 0 {
		remaining = append(remaining, o.Args...)
	}

	name := tempDirPrefix + "golang-http-" + seed
	o.AppName = name
	remaining = append(remaining, "-p", name)

	err = utils.RunCommandInteractive(o.Interactive, o.WorkDir, cmd, remaining...)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func (o *golangHTTPTest) thereShouldBeAJenkinsProjectCreate() error {
	fmt.Printf("TODO should be a jenkins project\n")
	return nil
}

func (o *golangHTTPTest) aRunningApplication() error {
	fmt.Printf("TODO should be able to query this using 'jx get app (app name)'\n")
	return nil
}

func (o *golangHTTPTest) executingJxDeleteApp() error {
	if !utils.DeleteApps() {
		return nil
	}
	appName := tempDirPrefix + "golang-http-" + seed
	cmd := "jx"
	fullAppName := o.GetGitOrganisation() + "/" + appName
	args := []string{"delete", "app", "-b", fullAppName}
	err := utils.RunCommandInteractive(o.Interactive, o.WorkDir, cmd, args...)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func (o *golangHTTPTest) executingJxDeleteRepo() error {
	if !utils.DeleteRepos() {
		return nil
	}
	appName := tempDirPrefix + "golang-http-" + seed
	cmd := "jx"
	args := []string{"delete", "repo", "-b", "--github", "-o", o.GetGitOrganisation(), "-n", appName}
	err := utils.RunCommandInteractive(o.Interactive, o.WorkDir, cmd, args...)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func GolangHTTPFeatureContext(s *godog.Suite) {
	o := &golangHTTPTest{
		Test: common.Test{
			Factory:     cmdutil.NewFactory(),
			Interactive: os.Getenv("JX_INTERACTIVE") == "true",
			Errors:      utils.CreateErrorSlice(),
		},
		Args: []string{},
	}

	s.BeforeSuite(func() {
		now := time.Now()
		seed = strconv.Itoa(int(now.Unix()))
	})
	s.AfterSuite(func() {
		os.RemoveAll(tmpDir)
	})

	s.Step(`^a work directory$`, o.aWorkDirectory)
	s.Step(`^running "([^"]*)" in that directory$`, o.runningInThatDirectory)
	s.Step(`^there should be a jenkins project created$`, o.thereShouldBeAJenkinsProjectCreate)
	s.Step(`^the application should be built and promoted via CI \/ CD$`, o.TheApplicationShouldBeBuiltAndPromotedViaCICD)
	s.Step(`^the application should be deleted after running jx delete app$`, o.executingJxDeleteApp)
	s.Step(`^the git repo should be deleted after running jx delete repo$`, o.executingJxDeleteRepo)
}

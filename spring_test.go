package godog_jx

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/common"
	"github.com/jenkins-x/godog-jx/utils"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
	"path/filepath"
)

type springTest struct {
	common.CommonTest

	Args []string
}

const (
	tempDirPrefix = "test-jx-create-spring-"
)

func (o *springTest) aWorkDirectory() error {
	tmpDir, err := ioutil.TempDir("", tempDirPrefix)
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

func (o *springTest) runningInThatDirectory(commandLine string) error {
	args := strings.Fields(commandLine)
	assert.NotEmpty(o.Errors, args, "not enough arguments")
	cmd := args[0]
	assert.Equal(o.Errors, "jx", cmd)
	gitProviderURL, err := o.GitProviderURL()
	if err != nil {
		return err
	}
	fmt.Printf("Using git provider URL %s and work directory %s\n", util.ColorInfo(gitProviderURL), util.ColorInfo(o.WorkDir))
	remaining := append(args[1:], "-b", "--git-provider-url", gitProviderURL)
	if len(o.Args) > 0 {
		remaining = append(remaining, o.Args...)
	}

	// add the artifact id using the current folder name
	_, name := filepath.Split(o.WorkDir)
	if strings.HasPrefix(name, tempDirPrefix) {
		name = "spring-" + strings.TrimPrefix(name, tempDirPrefix)
	}
	o.AppName = name
	remaining = append(remaining, "--artifact", name, "--name", name)

	err = utils.RunCommandInteractive(o.Interactive, o.WorkDir, cmd, remaining...)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func (o *springTest) thereShouldBeAJenkinsProjectCreate() error {
	fmt.Printf("TODO should be a jenkins project\n")
	return nil
}

func SpringFeatureContext(s *godog.Suite) {
	o := &springTest{
		CommonTest: common.CommonTest{
			Factory:     cmdutil.NewFactory(),
			Interactive: os.Getenv("JX_INTERACTIVE") == "true",
			Errors:      utils.CreateErrorSlice(),
		},
		Args: []string{},
	}
	s.Step(`^a work directory$`, o.aWorkDirectory)
	s.Step(`^running "([^"]*)" in that directory$`, o.runningInThatDirectory)
	s.Step(`^there should be a jenkins project created$`, o.thereShouldBeAJenkinsProjectCreate)
	s.Step(`^the application should be built and promoted via CI \/ CD$`, o.TheApplicationShouldBeBuiltAndPromotedViaCICD)
}

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
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
)

type importTest struct {
	common.Test

	SourceDir string
	Args      []string
}

func (o *importTest) aDirectoryContainingASpringBootApplication() error {
	tempDirPrefix := "import-"
	tmpDir, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return err
	}
	err = os.MkdirAll(tmpDir, utils.DefaultWritePermissions)
	if err != nil {
		return err
	}
	o.WorkDir = tmpDir
	assert.NotEmpty(o.Errors, o.SourceDir)
	assert.DirExists(o.Errors, o.SourceDir)
	err = utils.RunCommand("", "bash", "-c", "cp -r "+o.SourceDir+"/* "+o.WorkDir)
	if err != nil {
		return err
	}
	assert.DirExists(o.Errors, o.WorkDir)
	_, name := filepath.Split(o.WorkDir)
	o.AppName = name
	err = utils.ReplaceElement(filepath.Join(o.WorkDir, "pom.xml"), "artifactId", name, 1)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func (o *importTest) runningInThatDirectory(commandLine string) error {
	args := strings.Fields(commandLine)
	assert.NotEmpty(o.Errors, args, "not enough arguments")
	cmd := args[0]
	assert.Equal(o.Errors, "jx", cmd)
	gitProviderURL, err := o.GitProviderURL()
	if err != nil {
		return err
	}
	fmt.Printf("Using git provider URL %s\n", util.ColorInfo(gitProviderURL))
	remaining := append(args[1:], "-b", "--git-provider-url", gitProviderURL, "--org", o.GetGitOrganisation())
	if len(o.Args) > 0 {
		remaining = append(remaining, o.Args...)
	}
	err = utils.RunCommandInteractive(o.Interactive, o.WorkDir, cmd, remaining...)
	if err != nil {
		return err
	}
	return o.Errors.Error()
}

func (o *importTest) thereShouldBeAJenkinsProjectCreate() error {
	fmt.Printf("TODO should be a jenkins project\n")
	return nil
}

func ImportFeatureContext(s *godog.Suite) {
	o := &importTest{
		Test: common.Test{
			Factory:     cmdutil.NewFactory(),
			Interactive: os.Getenv("JX_INTERACTIVE") == "true",
			Errors:      utils.CreateErrorSlice(),
		},
		Args:      []string{},
		SourceDir: "examples/example-spring-boot",
	}
	s.Step(`^a directory containing a Spring Boot application$`, o.aDirectoryContainingASpringBootApplication)
	s.Step(`^running "([^"]*)" in that directory$`, o.runningInThatDirectory)
	//s.Step(`^there should be a jenkins project create$`, o.thereShouldBeAJenkinsProjectCreate)
	s.Step(`^there should be a jenkins project created$`, o.thereShouldBeAJenkinsProjectCreate)
	s.Step(`^the application should be built and promoted via CI \/ CD$`, o.TheApplicationShouldBeBuiltAndPromotedViaCICD)
}

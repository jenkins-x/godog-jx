package bitbucket

import (
	"fmt"
	"path/filepath"

	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/utils"
)

func (f *ForkFeature) thereIsNoForkOf(repo string) error {
	gitcmder := f.GitCommander
	err := gitcmder.DeleteWorkDir()
	if err != nil {
		return err
	}

	path := filepath.Join(f.GitCommander.Dir, repo)
	return AssertFileDoesNotExist(path)
}

func (f *ForkFeature) thereShouldBeAForkForTheCurrentUserWhichHasTheSameLastCommitAs(forkedRepo string) error {
	gitcmder := f.GitCommander
	upstreamSha, err := gitcmder.GetLastCommitSha(f.UpstreamDir)
	if err != nil {
		return err
	}
	forkSha, err := gitcmder.GetLastCommitSha(f.ForkDir)
	if err != nil {
		return err
	}
	utils.LogInfof("upstream last commit is %s\n", upstreamSha)
	utils.LogInfof("fork last commit is %s\n", forkSha)

	errors := CreateErrorSlice()
	assert := CreateAssert(errors)

	msg := fmt.Sprintf("The git sha on the fork should be the same as the upstream repository in dir %s and %s", f.ForkDir, f.UpstreamDir)
	assert.Equal(upstreamSha, forkSha, msg)
	return errors.Error()
}

func FeatureContext(s *godog.Suite) {
	f := &ForkFeature{
		GitCommander: CreateGitCommander(),
	}

	s.Step(`^there is no fork of "([^"]*)"$`, f.thereIsNoForkOf)
	s.Step(`^I fork the "([^"]*)" Bitbucket repo to the current user$`, f.iForkTheBitbucketRepoToTheCurrentUser)
	s.Step(`^there should be a fork for the current user which has the same last commit as "([^"]*)"$`, f.thereShouldBeAForkForTheCurrentUserWhichHasTheSameLastCommitAs)
}

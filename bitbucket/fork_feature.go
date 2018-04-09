package bitbucket

import (
	"time"

	"github.com/jenkins-x/godog-jx/utils"
)

type ForkFeature struct {
	GitCommander *GitCommander

	UpstreamDir    string
	ForkDir        string
	ForkedRepoName string
}

// ForkToUsersRepo forks the given upstream repo cleanly to the current github users account
func (f *ForkFeature) ForkToUsersRepo(uptreamRepoName string) (string, error) {
	gitcmder := f.GitCommander
	err := gitcmder.DeleteWorkDir()
	if err != nil {
		return "", err
	}
	err = f.iForkTheBitbucketRepoToTheCurrentUser(uptreamRepoName)
	if err != nil {
		return "", err
	}
	name := f.ForkedRepoName
	utils.LogInfof("forked the repository %s to the current users account %s\n", uptreamRepoName, name)
	return name, err
}

func (f *ForkFeature) iForkTheBitbucketRepoToTheCurrentUser(originalRepoName string) error {
	userRepo, err := ParseUserRepositoryName(originalRepoName)
	if err != nil {
		return err
	}
	currentBitbucketUser, err := utils.MandatoryEnvVar("BITBUCKET_USER")
	if err != nil {
		return err
	}
	f.ForkedRepoName = currentBitbucketUser + "/" + userRepo.Repository
	provider, err := CreateBitbucketCloudProvider()
	if err != nil {
		return err
	}
	gitcmder := f.GitCommander
	gitcmder.UseHttps = true

	upstreamRepo, err := GetRepository(provider, userRepo.Organisation, userRepo.Repository)
	if err != nil {
		return err
	}

	// now lets fork it
	repo, err := ForkRepositoryOrRevertMasterInFork(provider, userRepo, currentBitbucketUser)
	if err != nil {
		return err
	}

	utils.LogInfo("Waiting for fork to be available for cloning...")

	// Loop until repo is available or 30 seconds have passed
	for i := 0; i < 6; i++ {
		time.Sleep(5 * time.Second)
		_, err := GetRepository(provider, userRepo.Organisation, userRepo.Repository)

		if err == nil {
			utils.LogInfo("Fork is available for cloning!")
			break
		}
	}

	dir, err := gitcmder.Clone(repo)
	if err != nil {
		return err
	}
	utils.LogInfof("Cloned to directory: %s\n", dir)
	f.ForkDir = dir

	upstreamCloneURL, err := GetCloneURL(upstreamRepo, true)
	if err != nil {
		return err
	}

	upstreamDir, err := gitcmder.CloneFromURL(upstreamRepo, upstreamCloneURL)
	if err != nil {
		return err
	}
	f.UpstreamDir = upstreamDir

	err = gitcmder.ResetMasterFromUpstream(dir, upstreamCloneURL)
	if err != nil {
		return err
	}

	return nil
}

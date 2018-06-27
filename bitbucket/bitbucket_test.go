package bitbucket

import (
	"github.com/DATA-DOG/godog"
	"github.com/jenkins-x/godog-jx/utils"
	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/gits"
)

type BitbucketContext struct {
	GitCommander *GitCommander
	Provider     gits.GitProvider
	Owner        string
	Repo         string
	Head         string
	Base         string
}

type PullRequestFeature struct {
	Owner   string
	Repo    string
	Head    string
	Base    string
	Context *BitbucketContext
}

type ForkFeature struct {
	UpstreamDir    string
	ForkDir        string
	ForkedRepoName string
	Context        *BitbucketContext
}

func NewBitbucketContext() (*BitbucketContext, error) {

	user, err := utils.MandatoryEnvVar("BITBUCKET_USER")
	if err != nil {
		return nil, err
	}

	token, err := utils.MandatoryEnvVar("BITBUCKET_APP_PASSWORD")
	if err != nil {
		return nil, err
	}

	provider, _ := gits.NewBitbucketCloudProvider(
		&auth.AuthServer{
			URL:         "https://auth.example.com",
			Name:        "Test Auth Server",
			Kind:        "Oauth2",
			CurrentUser: "test-user",
		},
		&auth.UserAuth{
			Username: user,
			ApiToken: token,
		},
	)

	return &BitbucketContext{
		GitCommander: CreateGitCommander(),
		Provider:     provider,
	}, nil
}

func CreateForkFeature() (*ForkFeature, error) {
	bc, _ := NewBitbucketContext()

	return &ForkFeature{
		Context: bc,
	}, nil
}

func CreatePullRequestFeature() (*PullRequestFeature, error) {

	owner, err := utils.MandatoryEnvVar("BITBUCKET_OWNER")
	if err != nil {
		return nil, err
	}

	repo, err := utils.MandatoryEnvVar("BITBUCKET_REPO")
	if err != nil {
		return nil, err
	}

	head, err := utils.MandatoryEnvVar("BITBUCKET_HEAD")
	if err != nil {
		return nil, err
	}

	base, err := utils.MandatoryEnvVar("BITBUCKET_BASE")
	if err != nil {
		return nil, err
	}

	bc, _ := NewBitbucketContext()

	return &PullRequestFeature{
		Owner:   owner,
		Repo:    repo,
		Head:    head,
		Base:    base,
		Context: bc,
	}, nil
}

func FeatureContext(s *godog.Suite) {

	f, _ := CreateForkFeature()

	p, _ := CreatePullRequestFeature()

	s.Step(`^there is no fork of "([^"]*)"$`, f.thereIsNoForkOf)
	s.Step(`^I fork the "([^"]*)" Bitbucket repo to the current user$`, f.iForkTheBitbucketRepoToTheCurrentUser)
	s.Step(`^there should be a fork for the current user which has the same last commit as "([^"]*)"$`, f.thereShouldBeAForkForTheCurrentUserWhichHasTheSameLastCommitAs)

	s.Step(`^I have defined a valid Bitbucket organization, repo, and branches$`, p.iHaveDefinedAValidBitbucketOrganizationRepoAndBranches)
	s.Step(`^I create a pull request$`, p.iCreateAPullRequest)
	s.Step(`^there should be a pull request in the defined organization and repo$`, p.thereShouldBeAPullRequestInTheDefinedOrganizationAndRepo)
}

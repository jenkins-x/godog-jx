package bitbucket

import (
	"github.com/jenkins-x/godog-jx/utils"
	"github.com/jenkins-x/jx/pkg/gits"
)

func (p *PullRequestFeature) iHaveDefinedAValidBitbucketOrganizationRepoAndBranches() error {

	owner, err := utils.MandatoryEnvVar("BITBUCKET_OWNER")
	if err != nil {
		return err
	}

	repo, err := utils.MandatoryEnvVar("BITBUCKET_REPO")
	if err != nil {
		return err
	}

	_, err = p.Context.Provider.GetRepository(owner, repo)

	if err != nil {
		return err
	}

	return nil
}

func (p *PullRequestFeature) iCreateAPullRequest() error {

	owner, err := utils.MandatoryEnvVar("BITBUCKET_OWNER")
	if err != nil {
		return err
	}

	repo, err := utils.MandatoryEnvVar("BITBUCKET_REPO")
	if err != nil {
		return err
	}

	head, err := utils.MandatoryEnvVar("BITBUCKET_HEAD")
	if err != nil {
		return err
	}

	base, err := utils.MandatoryEnvVar("BITBUCKET_BASE")
	if err != nil {
		return err
	}

	prArgs := &gits.GitPullRequestArguments{
		Owner: owner,
		Repo:  repo,
		Title: "Test PR",
		Body:  "This PR is for testing purposes",
		Head:  head,
		Base:  base,
	}

	_, err = p.Context.Provider.CreatePullRequest(prArgs)

	if err != nil {
		return err
	}

	return nil
}

func (p *PullRequestFeature) thereShouldBeAPullRequestInTheDefinedOrganizationAndRepo() error {

	owner, err := utils.MandatoryEnvVar("BITBUCKET_OWNER")
	if err != nil {
		return err
	}

	repo, err := utils.MandatoryEnvVar("BITBUCKET_REPO")
	if err != nil {
		return err
	}

	_, err = p.Context.Provider.GetPullRequest(owner, repo, 1)

	if err != nil {
		return err
	}

	return nil
}

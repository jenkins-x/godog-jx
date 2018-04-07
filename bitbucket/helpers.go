package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jenkins-x/godog-jx/utils"
	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/gits"
	bitbucket "github.com/wbrefvem/go-bitbucket"
)

type UserRepositoryName struct {
	Organisation string
	Repository   string
}

func (r *UserRepositoryName) String() string {
	return r.Organisation + "/" + r.Repository
}

// ParseUserRepositoryName parses a repo name of the form `orgName/repoName` or returns a failure if
// the text cannot be parsed
func ParseUserRepositoryName(text string) (*UserRepositoryName, error) {
	values := strings.Split(text, "/")
	if len(values) != 2 {
		return nil, fmt.Errorf("Invalid github repository name. Expected the format `orgName/RepoName` but got %s", text)
	}
	return &UserRepositoryName{
		Organisation: values[0],
		Repository:   values[1],
	}, nil
}

// CreateBitbucketClient creates a new Bitbucket client
func CreateBitbucketCloudProvider() (*gits.BitbucketCloudProvider, error) {
	/*
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: "... your access token ..."},
		)
		tc := oauth2.NewClient(ctx, ts)
	*/

	user, err := utils.MandatoryEnvVar("BITBUCKET_USER")
	if err != nil {
		return nil, err
	}
	pwd, err := utils.MandatoryEnvVar("BITBUCKET_APP_PASSWORD")
	if err != nil {
		return nil, err
	}

	users := []*auth.UserAuth{
		{
			Username: user,
			ApiToken: pwd,
		},
	}
	authServer := auth.AuthServer{
		Name:  "Bitbucket Cloud",
		URL:   "https://bitbucket.org",
		Kind:  "Bitbucket Cloud",
		Users: users,
	}

	provider, err := gits.NewBitbucketCloudProvider(&authServer, users[0])
	bitbucketProvider, ok := provider.(*gits.BitbucketCloudProvider)

	if !ok {
		return nil, fmt.Errorf("BitbucketCloudProvider type assertion failed!")
	}
	return bitbucketProvider, nil
}

func GetRepository(provider *gits.BitbucketCloudProvider, owner string, name string) (*gits.GitRepository, error) {
	repo, err := provider.GetRepository(owner, name)
	return repo, err
}

// ForkRepositoryOrRevertMasterInFork forks the given repository to the new owner or resets the fork
// to the upstream master
func ForkRepositoryOrRevertMasterInFork(provider *gits.BitbucketCloudProvider, userRepo *UserRepositoryName, newOwner string) (*gits.GitRepository, error) {
	repoOwner := userRepo.Organisation
	repoName := userRepo.Repository
	repo, err := GetRepository(provider, repoOwner, repoName)
	if err != nil {
		return nil, err
	}
	u := repo.HTMLURL
	if u != "" {
		utils.LogInfof("Found repository at %s\n", u)
	}

	forkRepo, err := GetRepository(provider, newOwner, repoName)

	// TODO how to ignore just 404 errors?
	/*
		if err != nil {
			return nil, fmt.Errorf("Error checking if the fork already exists for %s/%s due to %#v", newOwner, repoName, err)
		}
	*/

	if forkRepo == nil || err != nil {
		utils.LogInfof("No fork available yet for %s/%s\n", newOwner, repoName)

		forkRepo, err = provider.ForkRepository(repoOwner, repoName, "")
		if err != nil {
			return nil, fmt.Errorf("Failed to fork repo %s to user %s due to %s", userRepo.String(), newOwner, err)
		}
	}
	return forkRepo, nil
}

// IsUser returns true if this name is a User or false if its an Organisation
func IsUser(client *bitbucket.APIClient, name string) (bool, error) {
	ctx := context.Background()
	user, _, err := client.UsersApi.UserGet(ctx)
	if err == nil && user.Username != "" {
		return true, nil
	}
	_, _, err = client.TeamsApi.TeamsUsernameGet(ctx, name)
	return false, err
}

func addAppPasswordToCloneUrl(cloneUrl string) string {
	parsedUrl, err := url.Parse(cloneUrl)

	if err != nil {
		return ""
	}
	parsedUrl.User = url.UserPassword(
		os.Getenv("BITBUCKET_USER"),
		os.Getenv("BITBUCKET_APP_PASSWORD"),
	)

	return parsedUrl.String()
}

func GetCloneURL(repo *gits.GitRepository, useHttps bool) (string, error) {
	cloneUrl := repo.SSHURL
	if useHttps {
		cloneUrl = addAppPasswordToCloneUrl(repo.HTMLURL)
		if cloneUrl == "" {
			return "", fmt.Errorf("Git repository does not have a clone URL: %v", repo)
		}
	} else {
		if cloneUrl == "" {
			return "", fmt.Errorf("Git repository does not have a SSH URL: %v", repo)
		}
	}
	return cloneUrl, nil
}

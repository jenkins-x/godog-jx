package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/jenkins-x/godog-jx/utils"
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

// CreateGitHubClient creates a new GitHub client
func CreateGitHubClient() (*github.Client, error) {
	/*
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: "... your access token ..."},
		)
		tc := oauth2.NewClient(ctx, ts)
	*/

	user, err := utils.MandatoryEnvVar("GITHUB_USER")
	if err != nil {
		return nil, err
	}
	pwd, err := utils.MandatoryEnvVar("GITHUB_PASSWORD")
	if err != nil {
		return nil, err
	}
	basicAuth := github.BasicAuthTransport{
		Username: user,
		Password: pwd,
	}
	httpClient := basicAuth.Client()
	return github.NewClient(httpClient), nil
}

func GetRepository(client *github.Client, owner string, name string) (*github.Repository, error) {
	ctx := context.Background()
	repo, _, err := client.Repositories.Get(ctx, owner, name)
	return repo, err
}

// ForkRepositoryOrRevertMasterInFork forks the given repository to the new owner or resets the fork
// to the upstream master
func ForkRepositoryOrRevertMasterInFork(client *github.Client, userRepo *UserRepositoryName, newOwner string) (*github.Repository, error) {
	repoOwner := userRepo.Organisation
	repoName := userRepo.Repository
	repo, err := GetRepository(client, repoOwner, repoName)
	if err != nil {
		return nil, err
	}
	u := repo.HTMLURL
	if u != nil {
		utils.LogInfof("Found repository at %s\n", *u)
	}

	forkRepo, err := GetRepository(client, newOwner, repoName)

	// TODO how to ignore just 404 errors?
	/*
	if err != nil {
		return nil, fmt.Errorf("Error checking if the fork already exists for %s/%s due to %#v", newOwner, repoName, err)
	}
	*/

	if forkRepo == nil || err != nil {
		utils.LogInfof("No fork available yet for %s/%s\n", newOwner, repoName)

		isUser, err := IsUser(client, repoOwner)
		if err != nil {
			return nil, err
		}
		opts := &github.RepositoryCreateForkOptions{}
		if !isUser {
			opts.Organization = newOwner
		}
		ctx := context.Background()
		forkRepo, _, err = client.Repositories.CreateFork(ctx, repoOwner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("Failed to fork repo %s to user %s due to %s", userRepo.String(), newOwner, err)
		}
	}
	return forkRepo, nil
}

// IsUser returns true if this name is a User or false if its an Organisation
func IsUser(client *github.Client, name string) (bool, error) {
	ctx := context.Background()
	user, _, err := client.Users.Get(ctx, name)
	if err == nil && user != nil {
		return true, nil
	}
	_, _, err = client.Organizations.Get(ctx, name)
	return false, err
}

func GetCloneURL(repo *github.Repository, useHttps bool) (string, error) {
	cloneUrl := repo.SSHURL
	if useHttps {
		cloneUrl = repo.CloneURL
		if cloneUrl == nil {
			return "", fmt.Errorf("Git repository does not have a clone URL: %v", repo)
		}
	} else {
		if cloneUrl == nil {
			return "", fmt.Errorf("Git repository does not have a SSH URL: %v", repo)
		}
	}
	return *cloneUrl, nil
}

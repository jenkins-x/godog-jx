package gits

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/util"
	"golang.org/x/oauth2"
)

type GitHubProvider struct {
	Username string
	Client   *github.Client
	Context  context.Context

	Server auth.AuthServer
	User   auth.UserAuth
}

func NewGitHubProvider(server *auth.AuthServer, user *auth.UserAuth) (GitProvider, error) {
	ctx := context.Background()

	provider := GitHubProvider{
		Server:   *server,
		User:     *user,
		Context:  ctx,
		Username: user.Username,
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: user.ApiToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	var err error
	u := server.URL
	if IsGitHubServerURL(u) {
		provider.Client = github.NewClient(tc)
	} else {
		u = GitHubEnterpriseApiEndpointURL(u)
		provider.Client, err = github.NewEnterpriseClient(u, u, tc)
	}
	return &provider, err
}

func GitHubEnterpriseApiEndpointURL(u string) string {
	// lets ensure we use the API endpoint to login
	if strings.Index(u, "/api/") < 0 {
		u = util.UrlJoin(u, "/api/v3/")
	}
	return u
}

// GetEnterpriseApiURL returns the github enterprise API URL or blank if this
// provider is for the https://github.com service
func (p *GitHubProvider) GetEnterpriseApiURL() string {
	u := p.Server.URL
	if IsGitHubServerURL(u) {
		return ""
	}
	return GitHubEnterpriseApiEndpointURL(u)
}

func IsGitHubServerURL(u string) bool {
	return u == "" || strings.HasPrefix(u, "https://github.com") || strings.HasPrefix(u, "github")
}

func (p *GitHubProvider) ListOrganisations() ([]GitOrganisation, error) {
	answer := []GitOrganisation{}
	pageSize := 100
	options := github.ListOptions{
		Page:    0,
		PerPage: pageSize,
	}
	for {
		orgs, _, err := p.Client.Organizations.List(p.Context, "", &options)
		if err != nil {
			return answer, err
		}

		for _, org := range orgs {
			name := org.Login
			if name != nil {
				o := GitOrganisation{
					Login: *name,
				}
				answer = append(answer, o)
			}
		}
		if len(orgs) < pageSize || len(orgs) == 0 {
			break
		}
		options.Page += 1
	}
	return answer, nil
}

func (p *GitHubProvider) ListRepositories(org string) ([]*GitRepository, error) {
	owner := org
	if owner == "" {
		owner = p.Username
	}
	answer := []*GitRepository{}
	pageSize := 100
	options := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: pageSize,
		},
	}
	for {
		repos, _, err := p.Client.Repositories.List(p.Context, owner, options)
		if err != nil {
			return answer, err
		}
		for _, repo := range repos {
			answer = append(answer, toGitHubRepo(asText(repo.Name), repo))
		}
		if len(repos) < pageSize || len(repos) == 0 {
			break
		}
		options.ListOptions.Page += 1
	}
	return answer, nil
}

func (p *GitHubProvider) GetRepository(org string, name string) (*GitRepository, error) {
	repo, _, err := p.Client.Repositories.Get(p.Context, org, name)
	if err != nil {
		return nil, fmt.Errorf("Failed to get repository %s/%s due to: %s", org, name, err)
	}
	return toGitHubRepo(name, repo), nil
}

func (p *GitHubProvider) CreateRepository(org string, name string, private bool) (*GitRepository, error) {
	repoConfig := &github.Repository{
		Name:    github.String(name),
		Private: github.Bool(private),
	}
	repo, _, err := p.Client.Repositories.Create(p.Context, org, repoConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create repository %s/%s due to: %s", org, name, err)
	}
	return toGitHubRepo(name, repo), nil
}

func (p *GitHubProvider) DeleteRepository(org string, name string) error {
	owner := org
	if owner == "" {
		owner = p.Username
	}
	_, err := p.Client.Repositories.Delete(p.Context, owner, name)
	if err != nil {
		return fmt.Errorf("Failed to delete repository %s/%s due to: %s", owner, name, err)
	}
	return err
}

func toGitHubRepo(name string, repo *github.Repository) *GitRepository {
	return &GitRepository{
		Name:             name,
		AllowMergeCommit: asBool(repo.AllowMergeCommit),
		CloneURL:         asText(repo.CloneURL),
		HTMLURL:          asText(repo.HTMLURL),
		SSHURL:           asText(repo.SSHURL),
		Fork:             asBool(repo.Fork),
		Language:         asText(repo.Language),
	}
}

func (p *GitHubProvider) ForkRepository(originalOrg string, name string, destinationOrg string) (*GitRepository, error) {
	repoConfig := &github.RepositoryCreateForkOptions{}
	if destinationOrg != "" {
		repoConfig.Organization = destinationOrg
	}
	repo, _, err := p.Client.Repositories.CreateFork(p.Context, originalOrg, name, repoConfig)
	if err != nil {
		msg := ""
		if destinationOrg != "" {
			msg = fmt.Sprintf(" to %s", destinationOrg)
		}
		owner := destinationOrg
		if owner == "" {
			owner = p.Username
		}
		if strings.Contains(err.Error(), "try again later") {
			fmt.Printf("Waiting for the fork of %s/%s to appear...\n", owner, name)
			// lets wait for the fork to occur...
			start := time.Now()
			deadline := start.Add(time.Minute)
			for {
				time.Sleep(5 * time.Second)
				repo, _, err = p.Client.Repositories.Get(p.Context, owner, name)
				if repo != nil && err == nil {
					break
				}
				t := time.Now()
				if t.After(deadline) {
					return nil, fmt.Errorf("Gave up waiting for Repository %s/%s to appear: %s", owner, name, err)
				}
			}
		} else {
			return nil, fmt.Errorf("Failed to fork repository %s/%s%s due to: %s", originalOrg, name, msg, err)
		}
	}
	answer := &GitRepository{
		Name:             name,
		AllowMergeCommit: asBool(repo.AllowMergeCommit),
		CloneURL:         asText(repo.CloneURL),
		HTMLURL:          asText(repo.HTMLURL),
		SSHURL:           asText(repo.SSHURL),
	}
	return answer, nil
}

func (p *GitHubProvider) CreateWebHook(data *GitWebHookArguments) error {
	owner := data.Owner
	if owner == "" {
		owner = p.Username
	}
	repo := data.Repo
	if repo == "" {
		return fmt.Errorf("Missing property Repo")
	}
	webhookUrl := data.URL
	if repo == "" {
		return fmt.Errorf("Missing property URL")
	}
	hooks, _, err := p.Client.Repositories.ListHooks(p.Context, owner, repo, nil)
	if err != nil {
		fmt.Printf("Error querying webhooks on %s/%s: %s\n", owner, repo, err)
	}
	for _, hook := range hooks {
		c := hook.Config["url"]
		s, ok := c.(string)
		if ok && s == webhookUrl {
			fmt.Printf("Already has a webhook registered for %s\n", webhookUrl)
			return nil
		}
	}
	config := map[string]interface{}{
		"url":          webhookUrl,
		"content_type": "json",
	}
	if data.Secret != "" {
		config["secret"] = data.Secret
	}
	hook := &github.Hook{
		Name:   github.String("web"),
		Config: config,
		Events: []string{"*"},
	}
	fmt.Printf("Creating github webhook for %s/%s for url %s\n", owner, repo, webhookUrl)
	_, _, err = p.Client.Repositories.CreateHook(p.Context, owner, repo, hook)
	return err
}

func (p *GitHubProvider) CreatePullRequest(data *GitPullRequestArguments) (*GitPullRequest, error) {
	owner := data.Owner
	repo := data.Repo
	title := data.Title
	body := data.Body
	head := data.Head
	base := data.Base
	config := &github.NewPullRequest{}
	if title != "" {
		config.Title = github.String(title)
	}
	if body != "" {
		config.Body = github.String(body)
	}
	if head != "" {
		config.Head = github.String(head)
	}
	if base != "" {
		config.Base = github.String(base)
	}
	pr, _, err := p.Client.PullRequests.Create(p.Context, owner, repo, config)
	if err != nil {
		return nil, err
	}
	return &GitPullRequest{
		URL:    notNullString(pr.HTMLURL),
		Owner:  owner,
		Repo:   repo,
		Number: pr.Number,
	}, nil
}

func (p *GitHubProvider) UpdatePullRequestStatus(pr *GitPullRequest) error {
	if pr.Number == nil {
		return fmt.Errorf("Missing Number for GitPullRequest %#v", pr)
	}
	n := *pr.Number
	result, _, err := p.Client.PullRequests.Get(p.Context, pr.Owner, pr.Repo, n)
	if err != nil {
		return err
	}
	head := result.Head
	if head != nil {
		pr.LastCommitSha = notNullString(head.SHA)
	} else {
		pr.LastCommitSha = ""
	}
	if result.Mergeable != nil {
		pr.Mergeable = result.Mergeable
	}
	pr.MergeCommitSHA = result.MergeCommitSHA
	if result.Merged != nil {
		pr.Merged = result.Merged
	}
	if result.ClosedAt != nil {
		pr.ClosedAt = result.ClosedAt
	}
	if result.MergedAt != nil {
		pr.MergedAt = result.MergedAt
	}
	if result.State != nil {
		pr.State = result.State
	}
	if result.StatusesURL != nil {
		pr.StatusesURL = result.StatusesURL
	}
	if result.IssueURL != nil {
		pr.IssueURL = result.IssueURL
	}
	if result.DiffURL != nil {
		pr.IssueURL = result.DiffURL
	}
	return nil
}

func (p *GitHubProvider) MergePullRequest(pr *GitPullRequest, message string) error {
	if pr.Number == nil {
		return fmt.Errorf("Missing Number for GitPullRequest %#v", pr)
	}
	n := *pr.Number
	ref := pr.LastCommitSha
	options := &github.PullRequestOptions{
		SHA: ref,
	}
	result, _, err := p.Client.PullRequests.Merge(p.Context, pr.Owner, pr.Repo, n, message, options)
	if err != nil {
		return err
	}
	if result.Merged == nil || *result.Merged == false {
		return fmt.Errorf("Failed to merge PR %s for ref %s as result did not return merged", pr.URL, ref)
	}
	return nil
}

func (p *GitHubProvider) AddPRComment(pr *GitPullRequest, comment string) error {
	if pr.Number == nil {
		return fmt.Errorf("Missing Number for GitPullRequest %#v", pr)
	}
	n := *pr.Number

	prComment := &github.IssueComment{
		Body: &comment,
	}
	_, _, err := p.Client.Issues.CreateComment(p.Context, pr.Owner, pr.Repo, n, prComment)
	if err != nil {
		return err
	}
	return nil
}

func (p *GitHubProvider) CreateIssueComment(owner string, repo string, number int, comment string) error {
	issueComment := &github.IssueComment{
		Body: &comment,
	}
	_, _, err := p.Client.Issues.CreateComment(p.Context, owner, repo, number, issueComment)
	if err != nil {
		return err
	}
	return nil
}

func (p *GitHubProvider) PullRequestLastCommitStatus(pr *GitPullRequest) (string, error) {
	ref := pr.LastCommitSha
	if ref == "" {
		return "", fmt.Errorf("Missing String for LastCommitSha %#v", pr)
	}
	results, _, err := p.Client.Repositories.ListStatuses(p.Context, pr.Owner, pr.Repo, ref, nil)
	if err != nil {
		return "", err
	}
	for _, result := range results {
		if result.State != nil {
			return *result.State, nil
		}
	}
	return "", fmt.Errorf("Could not find a status for repository %s/%s with ref %s", pr.Owner, pr.Repo, ref)
}

func (p *GitHubProvider) ListCommitStatus(org string, repo string, sha string) ([]*GitRepoStatus, error) {
	answer := []*GitRepoStatus{}
	results, _, err := p.Client.Repositories.ListStatuses(p.Context, org, repo, sha, nil)
	if err != nil {
		return answer, fmt.Errorf("Could not find a status for repository %s/%s with ref %s", org, repo, sha)
	}
	for _, result := range results {
		status := &GitRepoStatus{
			ID:          notNullInt64(result.ID),
			Context:     notNullString(result.Context),
			URL:         notNullString(result.URL),
			TargetURL:   notNullString(result.TargetURL),
			State:       notNullString(result.State),
			Description: notNullString(result.Description),
		}
		answer = append(answer, status)
	}
	return answer, nil
}

func notNullInt64(n *int64) int64 {
	if n != nil {
		return *n
	}
	return 0
}

func notNullString(tp *string) string {
	if tp == nil {
		return ""
	}
	return *tp
}

func (p *GitHubProvider) RenameRepository(org string, name string, newName string) (*GitRepository, error) {
	if org == "" {
		org = p.Username
	}
	config := &github.Repository{
		Name: github.String(newName),
	}
	repo, _, err := p.Client.Repositories.Edit(p.Context, org, name, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to edit repository %s/%s due to: %s", org, name, err)
	}
	answer := &GitRepository{
		Name:             name,
		AllowMergeCommit: asBool(repo.AllowMergeCommit),
		CloneURL:         asText(repo.CloneURL),
		HTMLURL:          asText(repo.HTMLURL),
		SSHURL:           asText(repo.SSHURL),
	}
	return answer, nil
}

func (p *GitHubProvider) ValidateRepositoryName(org string, name string) error {
	_, r, err := p.Client.Repositories.Get(p.Context, org, name)
	if err == nil {
		return fmt.Errorf("Repository %s already exists", GitRepoName(org, name))
	}
	if r.StatusCode == 404 {
		return nil
	}
	return err
}

func (p *GitHubProvider) UpdateRelease(owner string, repo string, tag string, releaseInfo *GitRelease) error {
	release := &github.RepositoryRelease{}
	rel, r, err := p.Client.Repositories.GetReleaseByTag(p.Context, owner, repo, tag)

	if r.StatusCode == 404 && !strings.HasPrefix(tag, "v") {
		// sometimes we prepend a v for example when using gh-release
		// so lets make sure we don't create a double release
		vtag := "v" + tag

		rel2, r2, err2 := p.Client.Repositories.GetReleaseByTag(p.Context, owner, repo, vtag)
		if r2.StatusCode != 404 {
			rel = rel2
			r = r2
			err = err2
			tag = vtag
		}
	}

	if r != nil && err == nil {
		release = rel
	}
	// lets populate the release
	if release.Name == nil && releaseInfo.Name != "" {
		release.Name = &releaseInfo.Name
	}
	if release.TagName == nil && releaseInfo.TagName != "" {
		release.TagName = &releaseInfo.TagName
	}
	if release.Body == nil && releaseInfo.Body != "" {
		release.Body = &releaseInfo.Body
	}
	if r.StatusCode == 404 {
		fmt.Printf("No release found for %s/%s and tag %s so creating a new release\n", owner, repo, tag)
		_, _, err = p.Client.Repositories.CreateRelease(p.Context, owner, repo, release)
		return err
	}
	id := release.ID
	if id == nil {
		return fmt.Errorf("The release for %s/%s tag %s has no ID!", owner, repo, tag)
	}
	r2, _, err := p.Client.Repositories.EditRelease(p.Context, owner, repo, *id, release)
	if r != nil {
		releaseInfo.URL = asText(r2.URL)
		releaseInfo.HTMLURL = asText(r2.HTMLURL)
	}
	return err
}

func (p *GitHubProvider) GetIssue(org string, name string, number int) (*GitIssue, error) {
	i, r, err := p.Client.Issues.Get(p.Context, org, name, number)
	if r.StatusCode == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p.fromGithubIssue(org, name, number, i)
}

func (p *GitHubProvider) SearchIssues(org string, name string, filter string) ([]*GitIssue, error) {
	opts := &github.IssueListByRepoOptions{}
	answer := []*GitIssue{}
	issues, r, err := p.Client.Issues.ListByRepo(p.Context, org, name, opts)
	if r.StatusCode == 404 {
		return answer, nil
	}
	if err != nil {
		return answer, err
	}
	for _, issue := range issues {
		if issue.Number != nil {
			n := *issue.Number
			i, err := p.fromGithubIssue(org, name, n, issue)
			if err != nil {
				return answer, err
			}

			// TODO apply the filter?
			answer = append(answer, i)
		}
	}
	return answer, nil
}

func (p *GitHubProvider) CreateIssue(owner string, repo string, issue *GitIssue) (*GitIssue, error) {
	labels := []string{}
	for _, label := range issue.Labels {
		name := label.Name
		if name != "" {
			labels = append(labels, name)
		}
	}
	config := &github.IssueRequest{
		Title:  &issue.Title,
		Body:   &issue.Body,
		Labels: &labels,
	}
	i, _, err := p.Client.Issues.Create(p.Context, owner, repo, config)
	if err != nil {
		return nil, err
	}
	number := 0
	if i.Number != nil {
		number = *i.Number
	}
	return p.fromGithubIssue(owner, repo, number, i)
}

func (p *GitHubProvider) fromGithubIssue(org string, name string, number int, i *github.Issue) (*GitIssue, error) {
	isPull := i.IsPullRequest()
	url := p.IssueURL(org, name, number, isPull)

	labels := []GitLabel{}
	for _, label := range i.Labels {
		labels = append(labels, toGitHubLabel(&label))
	}
	assignees := []GitUser{}
	for _, assignee := range i.Assignees {
		assignees = append(assignees, *toGitHubUser(assignee))
	}
	return &GitIssue{
		Number:        &number,
		URL:           url,
		State:         i.State,
		Title:         asText(i.Title),
		Body:          asText(i.Body),
		IsPullRequest: isPull,
		Labels:        labels,
		User:          toGitHubUser(i.User),
		ClosedBy:      toGitHubUser(i.ClosedBy),
		Assignees:     assignees,
	}, nil
}

func (p *GitHubProvider) IssueURL(org string, name string, number int, isPull bool) string {
	serverPrefix := p.Server.URL
	if !strings.HasPrefix(serverPrefix, "https://") {
		serverPrefix = "https://" + serverPrefix
	}
	path := "issues"
	if isPull {
		path = "pull"
	}
	url := util.UrlJoin(serverPrefix, org, name, path, strconv.Itoa(number))
	return url
}

func toGitHubUser(user *github.User) *GitUser {
	if user == nil {
		return nil
	}
	return &GitUser{
		Login:     asText(user.Login),
		Name:      asText(user.Name),
		Email:     asText(user.Email),
		AvatarURL: asText(user.AvatarURL),
	}
}

func toGitHubLabel(label *github.Label) GitLabel {
	return GitLabel{
		Name:  asText(label.Name),
		Color: asText(label.Color),
		URL:   asText(label.URL),
	}
}

func (p *GitHubProvider) HasIssues() bool {
	return true
}

func (p *GitHubProvider) IsGitHub() bool {
	return true
}

func (p *GitHubProvider) IsGitea() bool {
	return false
}

func (p *GitHubProvider) Kind() string {
	return "github"
}

func (p *GitHubProvider) JenkinsWebHookPath(gitURL string, secret string) string {
	return "/github-webhook/"
}

func GitHubAccessTokenURL(url string) string {
	if strings.Index(url, "://") < 0 {
		url = "https://" + url
	}
	return util.UrlJoin(url, "/settings/tokens/new?scopes=repo,read:user,user:email,write:repo_hook")
}

func (p *GitHubProvider) Label() string {
	return p.Server.Label()
}

func (p *GitHubProvider) ServerURL() string {
	return p.Server.URL
}

func asBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

func asText(text *string) string {
	if text != nil {
		return *text
	}
	return ""
}

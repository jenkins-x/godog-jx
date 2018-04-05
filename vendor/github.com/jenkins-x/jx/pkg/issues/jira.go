package issues

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/jenkins-x/jx/pkg/util"
)

type JiraService struct {
	JiraClient *jira.Client
	Server     *auth.AuthServer
	UserAuth   *auth.UserAuth
	Project    string
}

func CreateJiraIssueProvider(server *auth.AuthServer, userAuth *auth.UserAuth, project string) (IssueProvider, error) {
	if server.URL == "" {
		return nil, fmt.Errorf("No base URL for server!")
	}
	var httpClient *http.Client
	if userAuth != nil && !userAuth.IsInvalid() {
		user := userAuth.Username
		tp := jira.BasicAuthTransport{
			Username: user,
			Password: userAuth.ApiToken,
		}
		/*
		 */
		httpClient = tp.Client()
	}
	jiraClient, _ := jira.NewClient(httpClient, server.URL)
	return &JiraService{
		JiraClient: jiraClient,
		Server:     server,
		UserAuth:   userAuth,
		Project:    project,
	}, nil
}

func (i *JiraService) GetIssue(key string) (*gits.GitIssue, error) {
	issue, _, err := i.JiraClient.Issue.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return i.jiraToGitIssue(issue), nil
}

func (i *JiraService) SearchIssues(query string) ([]*gits.GitIssue, error) {
	jql := "project = " + i.Project + " AND status NOT IN (Closed, Resolved)"
	if query != "" {
		jql += " AND text ~ " + query
	}
	answer := []*gits.GitIssue{}
	issues, _, err := i.JiraClient.Issue.Search(jql, nil)
	if err != nil {
		return answer, err
	}
	for _, issue := range issues {
		answer = append(answer, i.jiraToGitIssue(&issue))
	}
	return answer, nil
}

func (i *JiraService) CreateIssue(issue *gits.GitIssue) (*gits.GitIssue, error) {
	project, _, err := i.JiraClient.Project.Get(i.Project)
	if err != nil {
		return nil, fmt.Errorf("Could not find project %s: %s", i.Project, err)
	}
	ji := i.gitToJiraIssue(issue)
	issueTypes := project.IssueTypes
	if len(issueTypes) > 0 {
		it := issueTypes[0]
		ji.Fields.Type.Name = it.Name
	}
	jira, resp, err := i.JiraClient.Issue.Create(ji)
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		msg := buf.String()
		return nil, fmt.Errorf("Failed to create issue: %s due to: %s", msg, err)
	}
	return i.jiraToGitIssue(jira), nil
}

func (i *JiraService) CreateIssueComment(key string, comment string) error {
	return fmt.Errorf("TODO")
}

func (i *JiraService) IssueURL(key string) string {
	return util.UrlJoin(i.Server.URL, "browse", key)
}

func (i *JiraService) jiraToGitIssue(issue *jira.Issue) *gits.GitIssue {
	answer := &gits.GitIssue{}
	key := issue.Key
	answer.Key = key
	answer.URL = i.IssueURL(key)
	fields := issue.Fields
	if fields != nil {
		answer.Title = fields.Summary
		answer.Body = fields.Description
		answer.Labels = gits.ToGitLabels(fields.Labels)
		answer.ClosedAt = jiraTimeToTimeP(fields.Resolutiondate)
		answer.User = jiraUserToGitUser(fields.Reporter)
		assignee := jiraUserToGitUser(fields.Assignee)
		if assignee != nil {
			answer.Assignees = []gits.GitUser{*assignee}
		}
	}
	return answer
}

func jiraUserToGitUser(user *jira.User) *gits.GitUser {
	if user == nil {
		return nil
	}
	return &gits.GitUser{
		AvatarURL: jiraAvatarUrl(user),
		Name:      user.Name,
		Login:     user.Key,
		Email:     user.EmailAddress,
	}
}
func jiraAvatarUrl(user *jira.User) string {
	answer := ""
	if user != nil {
		av := user.AvatarUrls
		answer = av.Four8X48
		if answer == "" {
			answer = av.Three2X32
		}
		if answer == "" {
			answer = av.Two4X24
		}
		if answer == "" {
			answer = av.One6X16
		}
	}
	return answer
}

func jiraTimeToTimeP(t jira.Time) *time.Time {
	tt := time.Time(t)
	return &tt
}

func (i *JiraService) gitToJiraIssue(issue *gits.GitIssue) *jira.Issue {
	answer := &jira.Issue{
		Fields: &jira.IssueFields{
			Project: jira.Project{
				Key: i.Project,
			},
			Summary:     issue.Title,
			Description: issue.Body,
			Type: jira.IssueType{
				Name: "Bug",
			},
		},
	}
	return answer
}

func (i *JiraService) ServerName() string {
	return i.Server.URL
}

func (i *JiraService) HomeURL() string {
	return util.UrlJoin(i.Server.URL, "browse", i.Project)
}

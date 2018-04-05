package quickstarts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/jenkins-x/jx/pkg/util"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	JenkinsXQuickstartsOwner = "jenkins-x-quickstarts"
)

// GitHubQuickstart returns a github based quickstart
func GitHubQuickstart(owner string, repo string, language string, framework string, tags ...string) *Quickstart {
	u := "https://github.com/" + owner + "/" + repo + "/archive/master.zip"

	return &Quickstart{
		ID:             owner + "/" + repo,
		Owner:          owner,
		Name:           repo,
		Language:       language,
		Framework:      framework,
		Tags:           tags,
		DownloadZipURL: u,
	}
}

func toGitHubQuickstart(owner string, repo *gits.GitRepository) *Quickstart {
	language := repo.Language
	// TODO find this from GitHub???
	framework := ""
	tags := []string{}
	return GitHubQuickstart(owner, repo.Name, language, framework, tags...)
}

func (m *QuickstartModel) LoadGithubQuickstarts(provider gits.GitProvider, owners []string) error {
	for _, owner := range owners {
		repos, err := provider.ListRepositories(owner)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			m.Add(toGitHubQuickstart(owner, repo))
		}
	}
	return nil
}

func NewQuickstartModel() *QuickstartModel {
	return &QuickstartModel{
		Quickstarts: map[string]*Quickstart{},
	}
}

// Add adds the given quickstart to this mode. Returns true if it was added
func (m *QuickstartModel) Add(q *Quickstart) bool {
	if q != nil {
		id := q.ID
		if id != "" {
			m.Quickstarts[id] = q
			return true
		}
	}
	return false
}

// CreateSurvey creates a survey to query pick a quickstart
func (model *QuickstartModel) CreateSurvey(filter *QuickstartFilter) (*QuickstartForm, error) {
	language := filter.Language
	if language != "" {
		languages := model.Languages()
		if len(languages) == 0 {
			// lets ignore this filter as there are none available
			filter.Language = ""
		} else {
			lower := strings.ToLower(language)
			lowerLanguages := util.StringArrayToLower(languages)
			if util.StringArrayIndex(lowerLanguages, lower) < 0 {
				return nil, util.InvalidOption("language", language, languages)
			}
		}
	}
	quickstarts := model.Filter(filter)
	names := []string{}
	m := map[string]*Quickstart{}
	for _, q := range quickstarts {
		name := q.SurveyName()
		m[name] = q
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		return nil, fmt.Errorf("No quickstarts match filter")
	}
	answer := ""
	if len(names) == 1 {
		answer = names[0]
	} else {
		prompt := &survey.Select{
			Message: "select the quickstart you wish to create",
			Options: names,
		}
		err := survey.AskOne(prompt, &answer, survey.Required)
		if err != nil {
			return nil, err
		}
	}
	if answer == "" {
		return nil, fmt.Errorf("No quickstart chosen")
	}
	q := m[answer]
	if q == nil {
		return nil, fmt.Errorf("Could not find chosen quickstart for %s", answer)
	}
	name, err := util.PickValue("Project name", q.Name, true)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("No project name")
	}
	form := &QuickstartForm{
		Quickstart: q,
		Name:       name,
	}
	return form, nil
}

// Filter filters all the available quickstarts with the filter and return the matches
func (model *QuickstartModel) Filter(filter *QuickstartFilter) []*Quickstart {
	answer := []*Quickstart{}
	for _, q := range model.Quickstarts {
		if filter.Matches(q) {
			answer = append(answer, q)
		}
	}
	return answer
}

// Languages returns all the languages in the quickstarts sorted
func (model *QuickstartModel) Languages() []string {
	m := map[string]string{}
	for _, q := range model.Quickstarts {
		l := q.Language
		if l != "" {
			m[l] = l
		}
	}
	return util.SortedMapKeys(m)
}

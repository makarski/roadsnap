package roadmap

import (
	"fmt"

	"github.com/andygrunwald/go-jira"

	"github.com/makarski/roadsnap/config"
)

type RoadmapViewer struct {
	jiraClient *jira.Client
}

func NewRoadmapViewer(cfg *config.JiraCrd) (*RoadmapViewer, error) {
	tp := jira.BasicAuthTransport{
		Username: cfg.User,
		Password: cfg.Token,
	}

	jiraClient, err := jira.NewClient(tp.Client(), cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init jira client: %s", err)
	}

	return &RoadmapViewer{jiraClient}, nil
}

func (rv *RoadmapViewer) ListEpics(project string) ([]jira.Issue, error) {
	// todo: add an option to set dates
	jql := fmt.Sprintf(`project="%s"&issuetype="Epic"&"Start date[Date]">startOfYear()`, project)

	epics, _, err := rv.jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch epics for project: %s. %s", project, err)
	}

	return epics, nil
}

func (rv *RoadmapViewer) ListEpicIssues(key string) ([]jira.Issue, error) {
	jql := fmt.Sprintf(`"Epic Link"=%s`, key)

	issues, _, err := rv.jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues for epic: %s. %s", key, err)
	}

	return issues, nil
}

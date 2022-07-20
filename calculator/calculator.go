package calculator

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/makarski/roadsnap/cmd/cache"
)

const (
	StatusDone       = "Done"
	StatusTodo       = "To Do"
	StatusInProgress = "In Progress"

	dateFormat = "January 2, 2006"
)

type Calculator struct {
	jiraBaseURL string
}

func NewCalculator(jiraBaseURL string) Calculator {
	return Calculator{jiraBaseURL}
}

func (c *Calculator) GenerateSummary(epics []*cache.EpicLink, project string) Summary {
	sum := Summary{
		Project:        project,
		epicLinkPrefix: c.jiraBaseURL + "/browse",
		Done:           make([]cache.EpicLink, 0),
		Overdue:        make([]cache.EpicLink, 0),
		Ongoing:        make([]cache.EpicLink, 0),
		Outstanding:    make([]cache.EpicLink, 0),
	}

	for _, epic := range epics {
		epic := *epic
		doneCnt, _, _ := epicStatusCount(epic)
		allDone := doneCnt == uint8(len(epic.Issues))
		dueDatePassed := epic.PastDueDate()

		if allDone && dueDatePassed {
			sum.Done = append(sum.Done, epic)
		}

		if !allDone && dueDatePassed {
			sum.Overdue = append(sum.Overdue, epic)
		}

		if epic.PreStartDate() {
			sum.Outstanding = append(sum.Outstanding, epic)
		}

		if epic.InProgress() {
			sum.Ongoing = append(sum.Ongoing, epic)
		}
	}

	return sum
}

func epicStatusCount(epic cache.EpicLink) (uint8, uint8, uint8) {
	var doneCount, inProgrCount, outstdCount uint8

	for _, issue := range epic.Issues {
		if isIssueDone(issue) {
			doneCount += 1
		} else if isIssueInProgress(issue) {
			inProgrCount += 1
		} else if isIssueToDo(issue) {
			outstdCount += 1
		}
	}

	return doneCount, inProgrCount, outstdCount
}

type (
	Summary struct {
		Project        string
		epicLinkPrefix string
		Done           []cache.EpicLink
		Overdue        []cache.EpicLink
		Ongoing        []cache.EpicLink
		Outstanding    []cache.EpicLink
	}

	NamedItems struct {
		Name  string
		Epics []cache.EpicLink
	}
)

func (s *Summary) NamedStats() []NamedItems {
	return []NamedItems{
		{
			Name:  "Done",
			Epics: s.Done,
		},
		{
			Name:  "Ongoing",
			Epics: s.Ongoing,
		},
		{
			Name:  "Overdue",
			Epics: s.Overdue,
		},
		{
			Name:  "Outstanding",
			Epics: s.Outstanding,
		},
	}
}

func (s *Summary) String() string {
	named := s.NamedStats()
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `
%s: %s
======================
`, s.Project, time.Now().Format(dateFormat))

	for _, item := range named {
		fmt.Fprintf(&buf, `
%s
----------------------
`, item.Name)

		for _, epic := range item.Epics {
			totalIssues := len(epic.Issues)
			doneCnt, inProgrCnt, outstdCnt := epicStatusCount(epic)

			labels := ""
			if len(epic.Epic.Fields.Labels) > 0 {
				labels = "`" + strings.Join(epic.Epic.Fields.Labels, "`, `") + "`"
			}

			fmt.Fprintf(&buf, `
#### %s [%s](%s/%s): %s
%s  
Start: %s  
Due: %s  
Total: %d, Done: %d, InProgress: %d, Outstanding: %d
`,
				quarterByDate(epic.DueDate),
				epic.Epic.Key,
				s.epicLinkPrefix,
				epic.Epic.Key,
				epic.Epic.Fields.Summary,
				labels,
				epic.StartDate.Format(dateFormat),
				epic.DueDate.Format(dateFormat),
				totalIssues, doneCnt, inProgrCnt, outstdCnt,
			)
		}
	}

	return buf.String()
}

func isIssueDone(issue jira.Issue) bool {
	return issue.Fields.Status.StatusCategory.Name == StatusDone
}

func isIssueToDo(issue jira.Issue) bool {
	return issue.Fields.Status.StatusCategory.Name == StatusTodo
}

func isIssueInProgress(issue jira.Issue) bool {
	return issue.Fields.Status.StatusCategory.Name == StatusInProgress
}

type Quarter int

const (
	Q1 Quarter = 1 + iota
	Q2
	Q3
	Q4
)

func (q Quarter) String() string {
	switch q {
	case Q1:
		return "Q1"
	case Q2:
		return "Q2"
	case Q3:
		return "Q3"
	case Q4:
		return "Q4"
	}

	return ""
}

func quarterByDate(date time.Time) Quarter {
	quarters := []Quarter{Q1, Q2, Q3, Q4}
	m := int(date.Month())

	for _, q := range quarters {
		qEnd := int(q * 3)
		qStart := int(qEnd - 2)

		if m <= qEnd && m >= qStart {
			return q
		}
	}

	return -1
}

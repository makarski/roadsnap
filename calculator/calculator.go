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
		project:        project,
		epicLinkPrefix: c.jiraBaseURL + "/browse",
		done:           make([]cache.EpicLink, 0),
		overdue:        make([]cache.EpicLink, 0),
		ongoing:        make([]cache.EpicLink, 0),
		outstanding:    make([]cache.EpicLink, 0),
	}

	for _, epic := range epics {
		epic := *epic
		doneCnt, _, _ := epicStatusCount(epic)
		allDone := doneCnt == uint8(len(epic.Issues))
		dueDatePassed := epic.PastDueDate()

		if allDone && dueDatePassed {
			sum.done = append(sum.done, epic)
		}

		if !allDone && dueDatePassed {
			sum.overdue = append(sum.overdue, epic)
		}

		if epic.PreStartDate() {
			sum.outstanding = append(sum.outstanding, epic)
		}

		if epic.InProgress() {
			sum.ongoing = append(sum.ongoing, epic)
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

type Summary struct {
	project        string
	epicLinkPrefix string
	done           []cache.EpicLink
	overdue        []cache.EpicLink
	ongoing        []cache.EpicLink
	outstanding    []cache.EpicLink
}

func (s *Summary) String() string {
	named := []struct {
		name  string
		epics []cache.EpicLink
	}{
		{
			name:  "Done",
			epics: s.done,
		},
		{
			name:  "Ongoing",
			epics: s.ongoing,
		},
		{
			name:  "Overdue",
			epics: s.overdue,
		},
		{
			name:  "Outstanding",
			epics: s.outstanding,
		},
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, `
%s: %s
======================
`, s.project, time.Now().Format(dateFormat))

	for _, item := range named {
		fmt.Fprintf(&buf, `
%s
----------------------
`, item.name)

		for _, epic := range item.epics {
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
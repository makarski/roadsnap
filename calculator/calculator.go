package calculator

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/makarski/roadsnap/cmd/cache"
	"github.com/makarski/roadsnap/config"
)

const (
	dateFormat = "January 2, 2006"
)

type Calculator struct {
	jiraBaseURL string
	statusNames *config.StatusNames
}

func NewCalculator(jiraBaseURL string, statusNames *config.StatusNames) Calculator {
	return Calculator{jiraBaseURL, statusNames}
}

func (c *Calculator) GenerateSummary(epics []*cache.EpicLink, project string, date time.Time) Summary {
	sum := Summary{
		Date:           date,
		Project:        project,
		epicLinkPrefix: c.jiraBaseURL + "/browse",
		Done:           make([]cache.EpicLink, 0),
		Overdue:        make([]cache.EpicLink, 0),
		Ongoing:        make([]cache.EpicLink, 0),
		Outstanding:    make([]cache.EpicLink, 0),

		statusConfigs: c.statusNames,
	}

	for _, epic := range epics {
		epic := *epic
		doneCnt, _, _ := statusCount(epic, c.statusNames)
		allDone := doneCnt == uint8(len(epic.Issues))
		dueDatePassed := epic.PastDueDate()

		epicStatusDone := isIssueDone(epic.Epic, c.statusNames)
		epicStatusToDo := isIssueToDo(epic.Epic, c.statusNames)

		if epicStatusDone && allDone {
			sum.Done = append(sum.Done, epic)
			continue
		}

		if dueDatePassed && (!epicStatusDone || !allDone) && !epicStatusToDo {
			sum.Overdue = append(sum.Overdue, epic)
			continue
		}

		if epicStatusToDo {
			sum.Outstanding = append(sum.Outstanding, epic)
			continue
		}

		if isIssueInProgress(epic.Epic, c.statusNames) {
			sum.Ongoing = append(sum.Ongoing, epic)
			continue
		}
	}

	return sum
}

func statusCount(epic cache.EpicLink, statusNames *config.StatusNames) (uint8, uint8, uint8) {
	counters := []*struct {
		statusConfigs []string
		counter       uint8
	}{
		{
			statusNames.Done,
			0,
		},
		{
			statusNames.InProgress,
			0,
		},
		{
			statusNames.ToDo,
			0,
		},
	}

	for _, issue := range epic.Issues {
		statusName := issue.Fields.Status.Name

		for _, counter := range counters {
			if sliceContains(counter.statusConfigs, statusName) {
				counter.counter += 1
			}
		}
	}

	return counters[0].counter, counters[1].counter, counters[2].counter
}

func isIssueDone(issue jira.Issue, statusConfig *config.StatusNames) bool {
	return sliceContains(statusConfig.Done, issue.Fields.Status.Name)
}

func isIssueToDo(issue jira.Issue, statusConfig *config.StatusNames) bool {
	return sliceContains(statusConfig.ToDo, issue.Fields.Status.Name)
}

func isIssueInProgress(issue jira.Issue, statusConfig *config.StatusNames) bool {
	return sliceContains(statusConfig.InProgress, issue.Fields.Status.Name)
}

func sliceContains(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}

	return false
}

type (
	Summary struct {
		Date           time.Time
		Project        string
		epicLinkPrefix string
		Done           []cache.EpicLink
		Overdue        []cache.EpicLink
		Ongoing        []cache.EpicLink
		Outstanding    []cache.EpicLink

		statusConfigs *config.StatusNames
	}

	NamedItems struct {
		Name  string
		Epics []cache.EpicLink
	}
)

func (s *Summary) AllCount() int {
	return len(s.Done) + len(s.Overdue) + len(s.Outstanding) + len(s.Ongoing)
}

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
			Name:  "To Do",
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
`, s.Project, s.Date.Format(dateFormat))

	for _, item := range named {
		fmt.Fprintf(&buf, `
%s (%d/%d)
----------------------
`, item.Name, len(item.Epics), s.AllCount())

		for _, epic := range item.Epics {
			totalIssues := len(epic.Issues)
			doneCnt, inProgrCnt, outstdCnt := statusCount(epic, s.statusConfigs)
			completeRatio := float64(doneCnt) / float64(totalIssues)

			labels := ""
			if len(epic.Epic.Fields.Labels) > 0 {
				labels = "`" + strings.Join(epic.Epic.Fields.Labels, "`, `") + "`"
			}

			statusAlert := ""
			if msg := epicStatusNotInSyncMessage(epic, s.statusConfigs); msg != "" {
				statusAlert = fmt.Sprintf(`
> %s
`, msg)
			}

			fmt.Fprintf(&buf, `
#### %s %d [%s](%s/%s): %s
%s
%s  
Status: %s  
Start: %s  
Due: %s  
Total: %d, Done: %d, InProgress: %d, Outstanding: %d  
Progress: %.2f
`,
				quarterByDate(epic.DueDate),
				epic.DueDate.Year(),
				epic.Epic.Key,
				s.epicLinkPrefix,
				epic.Epic.Key,
				epic.Epic.Fields.Summary,
				statusAlert,
				labels,
				epic.Epic.Fields.Status.Name,
				epic.StartDate.Format(dateFormat),
				epic.DueDate.Format(dateFormat),
				totalIssues, doneCnt, inProgrCnt, outstdCnt,
				completeRatio,
			)
		}
	}

	return buf.String()
}

func epicStatusNotInSyncMessage(epic cache.EpicLink, statusConfig *config.StatusNames) string {
	if (epic.PastDueDate() || epic.InActivePhase()) && isIssueToDo(epic.Epic, statusConfig) {
		return "Epic Status Does not correspond Planning Dates"
	}

	return ""
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

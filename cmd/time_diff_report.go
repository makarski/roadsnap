package cmd

import (
	"bytes"
	"fmt"
	"time"

	"github.com/makarski/roadsnap/calculator"
	"github.com/makarski/roadsnap/cmd/cache"
	"github.com/makarski/roadsnap/config"
	"github.com/makarski/roadsnap/util"
)

const viewDateFormat = "Jan 2, 2006"

func TimeWindowReport(cfg *config.Config) CmdFunc {
	statusConverter := calculator.NewStatusConverter(cfg.StatusNames)
	epicFinder := cache.NewEpicCacher(nil, InArgs.Dir)
	differ := calculator.NewTimeWindowDiffer(cfg.JiraCrd.BaseURL+"browse", statusConverter, epicFinder, InArgs.Dir)

	return func() error {
		year := time.Now().Year()

		for _, project := range cfg.Projects.Names {
			reports := make([]calculator.Report2, 0, 12)

			for i := 1; i <= 12; i++ {
				monthStart := time.Date(year, time.Month(i), 1, 0, 0, 0, 0, time.UTC)
				monthEnd := monthStart.AddDate(0, 1, -1)

				fmt.Println("> Generating report for", project, monthStart.Format("Jan, 2006"))

				report, err := differ.Report(project, monthStart, monthEnd)
				if err != nil {
					return fmt.Errorf("failed to build reports: %s", err)
				}

				reports = append(reports, *report)
			}

			filename := generateFileName(project, year)
			f, err := util.CreateFile(filename)
			if err != nil {
				return err
			}

			fmt.Fprint(f, ToMarkdown(project, reports))
			f.Close()
		}
		return nil
	}
}

// generateFileName returns a file name for the report
func generateFileName(project string, year int) string {
	project = util.RemoveSpaces(project)
	return fmt.Sprintf("%s/%s/%s-%d.md", InArgs.Dir, project, project, year)
}

func ToMarkdown(project string, reports []calculator.Report2) string {
	var overview, details bytes.Buffer

	overview.WriteString(fmt.Sprintf(`
%s: %s - %s
======

| Month | Snapshot From | Snapshot To | Progress | Epics Planned | Epics Done | Stories Planned | Stories Done |
| ---   | ---           | ---         | ---      | ---           | ---        | ---             | ---          |`,
		project, reports[0].From.Format("Jan, 2006"), reports[11].To.Format("Jan, 2006")))

	for _, report := range reports {
		mdO := fmt.Sprintf(`
| [%s](#%s) |%s | %s | %.2f | %d -> %d | %d -> %d | **%d** -> %d | %d -> **%d** |`,
			report.Title,
			report.To.Format("2006-01"),
			report.SnapshotFrom.Format(viewDateFormat),
			report.SnapshotTo.Format(viewDateFormat),
			report.Progress(),
			report.LeftEpicsPlanned,
			report.RightEpicsPlanned,
			report.LeftEpicsDone,
			report.RightEpicsDone,
			report.LeftStoriesPlanned,
			report.RightStoriesPlanned,
			report.LeftStoriesDone,
			report.RightStoriesDone,
		)

		overview.WriteString(mdO)

		mdD := fmt.Sprintf(`
---
<a name="%s"></a>%s
===

Snapshot From: %s  
Snapshot To: %s  
		
| Epic Name | Status | Planning | Due Date | Progress | Stories Total | Stories Done |
| ---       | ---    | ---      | ---      | ---      | ---		      | ---          |`,
			report.To.Format("2006-01"),
			report.Title,
			report.SnapshotFrom.Format(viewDateFormat),
			report.SnapshotTo.Format(viewDateFormat),
		)

		details.WriteString(mdD)

		for _, epicPair := range report.EpicPairs {
			mdD := fmt.Sprintf(`
| [%s](%s) %s | %s -> %s | %s | %s -> %s | %.2f | **%d** -> %d | %d -> **%d** |`,
				epicPair.Key,
				epicPair.Link,
				epicPair.Title,
				epicPair.Left.Status,
				epicPair.Right.Status,
				epicPair.PlanningStatus(),
				epicPair.LeftDueDate(viewDateFormat),
				epicPair.RightDueDate(viewDateFormat),
				epicPair.Progress(),
				len(epicPair.Left.PlanStories),
				len(epicPair.Right.PlanStories),
				epicPair.Left.StoriesDone,
				epicPair.Right.StoriesDone,
			)

			details.WriteString(mdD)
		}
	}

	details.WriteTo(&overview)

	return overview.String()
}

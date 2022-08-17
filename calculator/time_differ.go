package calculator

import (
	"fmt"
	"os"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/makarski/roadsnap/cmd/cache"
)

type (
	EpicFinder interface {
		FromCacheOrdered(time.Time, string) ([]*cache.EpicLink, error)
	}

	TimeWindowDiffer struct {
		linkPrefix      string
		statusConverter StatusConverter
		epicFinder      EpicFinder
		cacheDir        string
	}
)

func NewTimeWindowDiffer(linkPrefix string, statusConverter StatusConverter, epicFinder EpicFinder, cacheDir string) TimeWindowDiffer {
	return TimeWindowDiffer{
		linkPrefix:      linkPrefix,
		statusConverter: statusConverter,
		epicFinder:      epicFinder,
		cacheDir:        cacheDir,
	}
}

type historicDueDate struct {
	DueDate      time.Time
	ShapshotDate time.Time
	Key          string
}

func (twd *TimeWindowDiffer) findDueDates(
	snapshotDates *cache.CachedEntry,
	project string,
) (map[string][]historicDueDate, error) {
	fmt.Fprintln(os.Stderr, "> project cached:", snapshotDates)

	dueDates := make(map[string][]historicDueDate, 0)

	for _, date := range snapshotDates.Dates {
		date, err := time.Parse(cache.DateFormat, date)
		if err != nil {
			return nil, err
		}

		epics, err := twd.epicFinder.FromCacheOrdered(date, project)
		if err != nil {
			return nil, err
		}

		for _, epic := range epics {
			dueDates[epic.Epic.Key] = append(dueDates[epic.Epic.Key], historicDueDate{epic.DueDate, date, epic.Epic.Key})
		}
	}

	return dueDates, nil
}

func findSnapshotDatesForPeriod(dates []string, reportFrom, reportTo time.Time) (time.Time, time.Time, error) {
	var startSnapshot, endSnapshot, lastSnapshot string

	for _, snapDate := range dates {
		if snapDate >= reportFrom.Format(cache.DateFormat) &&
			(startSnapshot == "" || (startSnapshot != "" && snapDate <= startSnapshot)) {
			startSnapshot = snapDate
		}

		if snapDate <= reportTo.Format(cache.DateFormat) {
			endSnapshot = snapDate
		}

		lastSnapshot = snapDate
	}

	if startSnapshot == "" {
		startSnapshot = lastSnapshot
	}

	startSnapDate, err := time.Parse(cache.DateFormat, startSnapshot)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed parsing start snapdate: %s", err)
	}

	if endSnapshot == "" {
		return startSnapDate, startSnapDate, nil
	}

	endSnapDate, err := time.Parse(cache.DateFormat, endSnapshot)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed parsing end snapdate: %s", err)
	}

	return startSnapDate, endSnapDate, nil
}

func (twd *TimeWindowDiffer) Report(project string, reportFrom, reportTo time.Time) (*Report2, error) {
	projectsSnapshotDates, err := cache.ListSnapshotDates(twd.cacheDir, project)
	if err != nil {
		return nil, err
	}

	snapshotDates := projectsSnapshotDates[0]

	startSnapshotDate, endSnapshotDate, err := findSnapshotDatesForPeriod(
		snapshotDates.Dates,
		reportFrom,
		reportTo,
	)
	if err != nil {
		return nil, err
	}

	fromEpics, err := twd.epicFinder.FromCacheOrdered(startSnapshotDate, project)
	if err != nil {
		return nil, err
	}

	var toEpics []*cache.EpicLink
	if startSnapshotDate.Equal(endSnapshotDate) {
		toEpics = fromEpics
	} else {
		toEpics, err = twd.epicFinder.FromCacheOrdered(endSnapshotDate, project)
		if err != nil {
			return nil, err
		}
	}

	report := &Report2{
		Title:        reportFrom.Format("Jan, 2006"),
		From:         reportFrom,
		To:           reportTo,
		SnapshotFrom: startSnapshotDate,
		SnapshotTo:   endSnapshotDate,
	}

	twd.writeStats(report, fromEpics, toEpics)

	return report, nil
}

func (twd *TimeWindowDiffer) generateLink(key string) string {
	return twd.linkPrefix + "/" + key
}

func findAddPair(
	pairMap map[string]*Pair,
	pairs *[]*Pair,
	key string,
	item PlanEpic,
	left bool,
) {
	planningPair, ok := pairMap[key]
	if !ok {
		planningPair = &Pair{}
		pairMap[key] = planningPair
		*pairs = append(*pairs, planningPair)
	}

	planningPair.Key = item.Key
	planningPair.Link = item.Link
	planningPair.Title = item.Title

	if left {
		planningPair.Left = item
	} else {
		planningPair.Right = item
	}
}

func (twd *TimeWindowDiffer) writeToEpicPairs(
	report *Report2,
	epicMap map[string]*Pair,
	stateSlice []*cache.EpicLink,
	left bool,
) {
	for _, epicState := range stateSlice {
		epicState := *epicState
		if epicState.DueDate.Before(report.From) || epicState.DueDate.After(report.To) {
			continue
		}

		planEpic := twd.toPlanEpic(epicState)

		report.IncrPlanned(left, 1, len(epicState.Issues))
		report.IncrEpicDone(left, planEpic.Status, 1)

		for _, storyState := range epicState.Issues {
			planStory := twd.toPlanStory(epicState.SnapshotDate, storyState)
			(&planEpic).PlanStories = append(planEpic.PlanStories, &planStory)

			if planStory.Status.isDone() {
				(&planEpic).StoriesDone += 1
			}

			report.IncrStoriesDone(left, planStory.Status, 1)
		}

		findAddPair(epicMap, &report.EpicPairs, epicState.Epic.Key, planEpic, left)
	}
}

func (twd *TimeWindowDiffer) writeStats(report *Report2, fromState, toState []*cache.EpicLink) {
	epicPairsMap := make(map[string]*Pair, len(fromState))

	twd.writeToEpicPairs(
		report,
		epicPairsMap,
		fromState,
		true,
	)

	twd.writeToEpicPairs(
		report,
		epicPairsMap,
		toState,
		false,
	)
}

func (twd *TimeWindowDiffer) toPlanEpic(cached cache.EpicLink) PlanEpic {
	actualStatus := twd.statusConverter.Status(cached.Epic.Fields.Status.Name)

	return PlanEpic{
		Title:        cached.Epic.Fields.Summary,
		SnapshotDate: cached.SnapshotDate,
		StartDate:    cached.StartDate,
		DueDate:      cached.DueDate,
		Key:          cached.Epic.Key,
		Link:         twd.generateLink(cached.Epic.Key),
		Status:       actualStatus,
	}
}

func (twd *TimeWindowDiffer) toPlanStory(snapshotDate time.Time, jIssue jira.Issue) PlanStory {
	return PlanStory{
		SnapshotDate: snapshotDate,
		Key:          jIssue.Key,
		Title:        jIssue.Fields.Summary,
		Link:         twd.generateLink(jIssue.Key),
		Status:       twd.statusConverter.Status(jIssue.Fields.Status.Name),
	}
}

type (
	Report2 struct {
		Title        string
		From         time.Time
		To           time.Time
		SnapshotFrom time.Time
		SnapshotTo   time.Time

		LeftEpicsPlanned   int
		LeftEpicsDone      int
		LeftStoriesPlanned int
		LeftStoriesDone    int

		RightEpicsPlanned   int
		RightEpicsDone      int
		RightStoriesPlanned int
		RightStoriesDone    int

		EpicPairs []*Pair
	}

	Pair struct {
		Key   string
		Title string
		Link  string

		Left  PlanEpic
		Right PlanEpic
	}

	PlanEpic struct {
		Title        string
		SnapshotDate time.Time
		StartDate    time.Time
		DueDate      time.Time
		Key          string
		Link         string
		Status       Status
		PlanStories  []*PlanStory
		StoriesDone  int
	}

	PlanStory struct {
		SnapshotDate time.Time
		Key          string
		Title        string
		Link         string
		Status       Status
	}
)

func (p *Pair) Progress() float64 {
	if p.hasLeft() && p.hasRight() && p.Right.StoriesDone > 0 {
		return float64(p.Right.StoriesDone) / float64(len(p.Left.PlanStories))
	}

	if !p.hasLeft() && p.hasRight() && p.Right.StoriesDone > 0 {
		return float64(p.Right.StoriesDone) / float64(len(p.Right.PlanStories))
	}

	return 0
}

func (p *Pair) hasLeft() bool {
	return p.Left.Title != ""
}

func (p *Pair) hasRight() bool {
	return p.Right.Title != ""
}

func (p *Pair) PlanningStatus() PlanningStatus {
	if p.hasLeft() && p.hasRight() {
		if p.Left.DueDate.Before(p.Right.DueDate) {
			return PlanningStatusPostponed
		}

		if p.Left.DueDate.After(p.Right.DueDate) {
			return PlanningStatusAdvanced
		}

		return PlanningStatusOK
	}

	return PlanningStatusReplanned
}

func (p *Pair) LeftDueDate(format string) string {
	return p.leftRightDueDate(format, true)
}

func (p *Pair) RightDueDate(format string) string {
	return p.leftRightDueDate(format, false)
}

func (p *Pair) leftRightDueDate(format string, left bool) string {
	if left && p.hasLeft() {
		return p.Left.DueDate.Format(format)
	}

	if !left && p.hasRight() {
		return p.Right.DueDate.Format(format)
	}

	return "Rescheduled"
}

func (r *Report2) Progress() float64 {
	if r.RightStoriesDone == 0 {
		return 0
	}

	return float64(r.RightStoriesDone) / float64(r.LeftStoriesPlanned)
}

func (r *Report2) IncrPlanned(left bool, epicCount, storyCount int) {
	if left {
		r.LeftEpicsPlanned += epicCount
		r.LeftStoriesPlanned += storyCount
	} else {
		r.RightEpicsPlanned += epicCount
		r.RightStoriesPlanned += storyCount
	}
}

func (r *Report2) IncrEpicDone(left bool, status Status, count int) {
	if !status.isDone() {
		return
	}

	if left {
		r.LeftEpicsDone += count
	} else {
		r.RightEpicsDone += count
	}
}

func (r *Report2) IncrStoriesDone(left bool, status Status, count int) {
	if !status.isDone() {
		return
	}

	if left {
		r.LeftStoriesDone += count
	} else {
		r.RightStoriesDone += count
	}
}

package cache

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/makarski/roadsnap/roadmap"

	"github.com/andygrunwald/go-jira"
)

const dateFormat = "2006-01-02"

var CustomFieldStartDate = ""

type EpicCacher struct {
	rv           *roadmap.RoadmapViewer
	baseDir      string
	projCacheDir func(string, time.Time) string
}

func NewEpicCacher(rv *roadmap.RoadmapViewer, dir string) *EpicCacher {
	return &EpicCacher{
		rv,
		dir,
		func(project string, snapshotDate time.Time) string {
			return path.Join(dir, removeSpace(project), snapshotDate.Format(dateFormat), "raw_data")
		},
	}
}

func (ec *EpicCacher) cacheNameEpic(date time.Time, project string) string {
	projectKey := removeSpace(project)

	epicFileKey := fmt.Sprintf("epics_%s.json", projectKey)
	return path.Join(ec.projCacheDir(project, date), projectKey, epicFileKey)
}

func (ec *EpicCacher) cacheNameIssues(date time.Time, project, epicKey string) string {
	return path.Join(ec.projCacheDir(project, date), removeSpace(project), fmt.Sprintf("issues_%s.json", epicKey))
}

func removeSpace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

func createDir(key string) error {
	if err := os.MkdirAll(path.Dir(key), os.FileMode(0744)); err != nil {
		return fmt.Errorf("failed to create dir for key: %s. %s", key, err)
	}
	return nil
}

func (ec *EpicCacher) Cache(date time.Time, projects []string) error {
	for _, projectName := range projects {
		epics, err := ec.cacheEpics(date, projectName)
		if err != nil {
			return err
		}

		for _, epic := range epics {
			issues, err := ec.cacheEpicIssues(date, projectName, epic.Key)
			if err != nil {
				return err
			}

			fmt.Printf("> cached %d issues for epic: %s\n", len(issues), epic.Fields.Summary)
		}
	}

	return nil
}

type EpicLink struct {
	SnapshotDate time.Time
	StartDate    time.Time
	DueDate      time.Time
	Epic         jira.Issue
	Issues       []jira.Issue
}

func (el *EpicLink) PastDueDate() bool {
	return el.SnapshotDate.After(el.DueDate)
}

func (el *EpicLink) PreStartDate() bool {
	return el.SnapshotDate.Before(el.StartDate)
}

func (el *EpicLink) InProgress() bool {
	return el.SnapshotDate.After(el.StartDate) && el.SnapshotDate.Before(el.DueDate)
}

func (el *EpicLink) UnmarshalJSON(b []byte) error {
	errFmt := "EpicLink.UnmarshalJSON: %s"

	var epicIssue jira.Issue
	if err := json.Unmarshal(b, &epicIssue); err != nil {
		return fmt.Errorf(errFmt, fmt.Sprintf("failed to unmarshal jira issue: %s", err))
	}

	el.Epic = epicIssue
	el.DueDate = time.Time(epicIssue.Fields.Duedate)

	// return is custom field for start date not defined
	if CustomFieldStartDate == "" {
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf(errFmt, fmt.Sprintf("failed to unmarshal raw object: %s", err))
	}

	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(raw["fields"], &rawFields); err != nil {
		return fmt.Errorf(errFmt, fmt.Sprintf("failed to unmarshal rawFields: %s", err))
	}

	var startDateStr string
	if err := json.Unmarshal(rawFields[CustomFieldStartDate], &startDateStr); err != nil {
		return fmt.Errorf(errFmt, fmt.Sprintf("failed to unmarshal `%s`: %s", CustomFieldStartDate, err))
	}

	sd, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fmt.Errorf(errFmt, fmt.Sprintf("failed to parse start time: %s", err))
	}

	el.StartDate = sd

	return nil
}

func (ec *EpicCacher) FromCache(date time.Time, projectName string) ([]*EpicLink, error) {
	f, err := os.Open(ec.cacheNameEpic(date, projectName))
	defer f.Close()
	if err != nil {
		return nil, err
	}

	var epicLinks []*EpicLink
	if err := json.NewDecoder(f).Decode(&epicLinks); err != nil {
		return nil, err
	}

	for _, epic := range epicLinks {
		f, err := os.Open(ec.cacheNameIssues(date, projectName, epic.Epic.Key))
		if err != nil {
			return nil, fmt.Errorf("failed to read cached issues for epic: %s. %s", epic.Epic.Key, err)
		}

		var issues []jira.Issue
		if err := json.NewDecoder(f).Decode(&issues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal issues for epic: %s. %s", epic.Epic.Key, err)
		}

		f.Close()

		epic.Issues = issues
		epic.SnapshotDate = date
	}

	return epicLinks, nil
}

func (ec *EpicCacher) cacheEpics(date time.Time, projectName string) ([]jira.Issue, error) {
	epics, err := ec.rv.ListEpics(projectName)
	if err != nil {
		return nil, err
	}

	fileKey := ec.cacheNameEpic(date, projectName)
	f, err := createFile(fileKey)
	if err != nil {
		return nil, err
	}

	err = json.NewEncoder(f).Encode(epics)
	return epics, err
}

func (ec *EpicCacher) cacheEpicIssues(date time.Time, project, key string) ([]jira.Issue, error) {
	issues, err := ec.rv.ListEpicIssues(key)
	if err != nil {
		return nil, err
	}

	fileKey := ec.cacheNameIssues(date, project, key)
	f, err := createFile(fileKey)
	if err != nil {
		return nil, err
	}

	err = json.NewEncoder(f).Encode(issues)
	return issues, err
}

func createFile(fileKey string) (*os.File, error) {
	if err := createDir(fileKey); err != nil {
		return nil, err
	}

	f, err := os.Create(fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file for key: %s. %s", fileKey, err)
	}

	return f, nil
}

type CachedEntry struct {
	Project string
	Dates   []string
}

func ListProjects(baseDir string) ([]*CachedEntry, error) {
	records := make([]*CachedEntry, 0)
	byProject := make(map[string]*CachedEntry)

	fs.WalkDir(os.DirFS(baseDir), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		match, err := path.Match("*/*/raw_data", p)
		if err != nil {
			return err
		}

		if !match {
			return nil
		}

		pathData := strings.SplitN(path.Dir(p), "/", 2)

		if bp, ok := byProject[pathData[0]]; ok {
			bp.Dates = append(bp.Dates, pathData[1])
		} else {
			record := &CachedEntry{pathData[0], []string{pathData[1]}}
			byProject[pathData[0]] = record
			records = append(records, record)
		}

		return nil
	})

	return records, nil
}

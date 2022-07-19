package list

import (
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"github.com/makarski/roadsnap/calculator"
	"github.com/makarski/roadsnap/cmd/cache"
)

const (
	StatusDone       = "Done"
	StatusTodo       = "To Do"
	StatusInProgress = "In Progress"
)

type (
	CacheReader interface {
		FromCache(time.Time, string) ([]*cache.EpicLink, error)
	}

	SummaryGenerator interface {
		GenerateSummary([]*cache.EpicLink, string) calculator.Summary
	}
)

type Lister struct {
	cr        CacheReader
	sg        SummaryGenerator
	targetDir string
}

func NewLister(cr CacheReader, sg SummaryGenerator, targetDir string) *Lister {
	return &Lister{cr, sg, targetDir}
}

func (l *Lister) List(date time.Time, project string) error {
	epics, err := l.cr.FromCache(date, project)
	if err != nil {
		return err
	}

	sort.Slice(epics, func(i, j int) bool {
		return epics[i].DueDate.Unix() < epics[j].DueDate.Unix()
	})

	summary := l.sg.GenerateSummary(epics, project)
	reportTxt := summary.String()

	fileKey := path.Join(l.targetDir, project, date.Format(cache.DateFormat), project+"_roadsnap.md")

	f, err := os.Create(fileKey)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("failed to created report file: %s. %s", fileKey, err)
	}

	_, err = f.WriteString(reportTxt)
	return err
}

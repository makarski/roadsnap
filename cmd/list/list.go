package list

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/makarski/roadsnap/calculator"
	"github.com/makarski/roadsnap/cmd/cache"
)

type (
	CacheReader interface {
		FromCacheOrdered(time.Time, string) ([]*cache.EpicLink, error)
	}

	SummaryGenerator interface {
		GenerateSummary([]*cache.EpicLink, string, time.Time) calculator.Summary
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

func (l *Lister) WriteReport(date time.Time, project string) error {
	summary, err := l.GenerateSummary(date, project)
	if err != nil {
		return err
	}

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

func (l *Lister) GenerateSummary(date time.Time, project string) (calculator.Summary, error) {
	epics, err := l.cr.FromCacheOrdered(date, project)
	if err != nil {
		return calculator.Summary{}, err
	}

	return l.sg.GenerateSummary(epics, project, date), nil
}

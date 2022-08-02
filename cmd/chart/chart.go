package chart

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"github.com/makarski/roadsnap/calculator"
)

type (
	SummaryGenerator interface {
		GenerateSummary(time.Time, string) (calculator.Summary, error)
	}

	ByDate struct {
		Date  time.Time
		Stats []DataItem
	}

	DataItem struct {
		Name     string
		Value    int
		MaxValue int
	}
)

type Drawer struct {
	sg  SummaryGenerator
	dir string
}

func NewDrawer(sg SummaryGenerator, dir string) Drawer {
	return Drawer{sg, dir}
}

func (d *Drawer) Draw(dates []time.Time, project string) error {
	errFmt := "ChartGenerator.Draw: %s"

	byDate := make([]*ByDate, 0)
	byDateMap := make(map[time.Time]*ByDate, 0)

	for _, date := range dates {
		summary, err := d.sg.GenerateSummary(date, project)
		if err != nil {
			return fmt.Errorf(errFmt, fmt.Sprintf("failed to generate summary for %s, %s. %s", date, project, err))
		}

		for _, stat := range summary.NamedStats() {
			byDateItem, ok := byDateMap[date]
			if !ok {
				byDateItem = &ByDate{Date: date}
				byDateMap[date] = byDateItem
				byDate = append(byDate, byDateItem)
			}

			byDateItem.Stats = append(byDateItem.Stats, DataItem{Name: stat.Name, Value: len(stat.Epics), MaxValue: summary.AllCount()})
		}
	}

	chart.DefaultBackgroundColor = chart.ColorTransparent
	chart.DefaultCanvasColor = chart.ColorTransparent

	barWidth := 150

	stackedBarChart := chart.StackedBarChart{
		Title:      project,
		TitleStyle: chart.StyleTextDefaults(),
		Background: chart.Style{
			Padding: chart.Box{
				Top:    100,
				Bottom: 20,
			},
		},
		Width:      810,
		Height:     500,
		XAxis:      chart.StyleTextDefaults(),
		YAxis:      chart.StyleTextDefaults(),
		BarSpacing: 50,
	}

	bars := make([]chart.StackedBar, 0, len(byDate))

	for _, bd := range byDate {
		bar := &chart.StackedBar{
			Name:  bd.Date.Format("Jan 02, 2006"),
			Width: barWidth,
		}

		for _, stat := range bd.Stats {
			color := colorByName(stat.Name)
			barVal := chart.Value{
				Label: fmt.Sprintf("%s (%d/%d)", stat.Name, stat.Value, stat.MaxValue),
				Value: float64(stat.Value),
				Style: chart.Style{
					StrokeWidth: .01,
					FillColor:   color,
					FontColor:   drawing.ColorWhite,
				},
			}

			bar.Values = append(bar.Values, barVal)
		}

		bars = append(bars, *bar)
	}

	stackedBarChart.Bars = bars

	f, err := os.Create(path.Join(d.dir, project, "roadmap-stats.png"))
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	if err := stackedBarChart.Render(chart.PNG, f); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	return f.Close()
}

func colorByName(name string) drawing.Color {
	switch name {
	case "Done":
		return drawing.ColorGreen
	case "To Do":
		return drawing.Color{R: 100, G: 80, B: 90, A: 255}
	case "Overdue":
		return drawing.ColorRed
	}

	return drawing.ColorBlue
}

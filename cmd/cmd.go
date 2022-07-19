package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/makarski/roadsnap/cmd/cache"
	"github.com/makarski/roadsnap/cmd/list"
	"github.com/makarski/roadsnap/config"
	"github.com/makarski/roadsnap/roadmap"
)

const dateFormat = "2006-01-02"

type (
	CmdFunc   = func() error
	CmdRunner = func(*config.Config) CmdFunc

	Flags struct {
		Dir         string
		Interactive bool
		ConfigFile  string
	}
)

var (
	InArgs = Flags{}

	cmds = map[string]CmdRunner{
		"cache": cacheCmd,
		"list":  listCmd,
	}

	out         = os.Stdout
	interactOut = os.Stderr
	in          = os.Stdin
)

func cacheCmd(cfg *config.Config) CmdFunc {
	snapshotDate := time.Now()

	return func() error {
		rv, err := roadmap.NewRoadmapViewer(cfg.JiraCrd)
		if err != nil {
			return err
		}

		cacher := cache.NewEpicCacher(rv, InArgs.Dir)

		// cache all project and return
		if !InArgs.Interactive {
			fmt.Fprintln(out, "> Caching projects:\n  *", strings.Join(cfg.Projects.Names, "\n  * "))
			return cacher.Cache(snapshotDate, cfg.Projects.Names)
		}

		// pick a project interactively
		for i, projectName := range cfg.Projects.Names {
			fmt.Fprintf(interactOut, "  * %d: %s\n", i, projectName)
		}

		fmt.Fprintf(interactOut, "\n> Pick a project to cache (ex: 2): ")

		var pPick int
		if _, err := fmt.Fscanf(in, "%d\n", &pPick); err != nil {
			return err
		}

		project := cfg.Projects.Names[pPick]

		fmt.Fprintln(out, "> Caching project:", project)

		return cacher.Cache(snapshotDate, []string{project})
	}
}

func listCmd(cfg *config.Config) CmdFunc {
	cache.CustomFieldStartDate = cfg.Epic.CustomFieldStartDate

	return func() error {
		projects, err := cache.ListProjects(InArgs.Dir)
		if err != nil {
			return err
		}

		cacheReader := cache.NewEpicCacher(nil, InArgs.Dir)
		lister := list.NewLister(cacheReader, cfg.JiraCrd.BaseURL, InArgs.Dir)

		if InArgs.Interactive {
			return interactListCmdHandler(lister, projects)
		}

		for _, project := range projects {
			if len(project.Dates) == 0 {
				fmt.Fprintf(out, "> Skipping project '%s' - no cached raw data\n", project.Project)
				continue
			}

			sort.Sort(sort.Reverse(sort.StringSlice(project.Dates)))

			// only process the last one
			t, err := time.Parse(dateFormat, project.Dates[0])
			if err != nil {
				return fmt.Errorf("failed to parse time for project: %s. %s", project.Project, err)
			}

			if err := lister.List(t, project.Project); err != nil {
				return fmt.Errorf("failed to list project: %s. %s", project.Project, err)
			}
		}

		return nil
	}
}

func interactListCmdHandler(lister *list.Lister, projects []*cache.CachedEntry) error {
	for i, project := range projects {
		fmt.Fprintf(interactOut, "\n  * %d: %s\n", i, project.Project)

		for j, date := range project.Dates {
			fmt.Fprintf(interactOut, "  | - %d: %s\n", j, date)
		}
	}

	fmt.Fprintf(interactOut, "\n> Enter project and date index (ex: 2, 0): ")

	var pPick, dPick int
	if _, err := fmt.Fscanf(in, "%d, %d\n", &pPick, &dPick); err != nil {
		return err
	}

	project := projects[pPick]
	date := project.Dates[dPick]

	t, err := time.Parse(dateFormat, date)
	if err != nil {
		return err
	}

	return lister.List(t, project.Project)
}

func Run(cmdName string) error {
	cfg, err := config.LoadConfig(InArgs.ConfigFile)
	if err != nil {
		return err
	}

	cmdRun, ok := cmds[cmdName]
	if !ok {
		return fmt.Errorf("command %s not defined", cmdName)
	}

	return cmdRun(cfg)()
}
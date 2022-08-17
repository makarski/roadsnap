package calculator

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/makarski/roadsnap/config"
)

type (
	Status         string
	PlanningStatus string
)

func (s Status) isDone() bool       { return s == StatusDone }
func (s Status) isToDo() bool       { return s == StatusToDo }
func (s Status) isInProgress() bool { return s == StatusInProgress }

const (
	StatusDone       Status = "Done"
	StatusToDo       Status = "ToDo"
	StatusInProgress Status = "InProgress"
	StatusUndefined  Status = "undefined"

	PlanningStatusOverdue   PlanningStatus = "Overdue"
	PlanningStatusPostponed PlanningStatus = "Postponed"
	PlanningStatusAdvanced  PlanningStatus = "Advanced"
	PlanningStatusOK        PlanningStatus = "Ok"
	PlanningStatusReplanned PlanningStatus = "Replanned"
)

type StatusConverter struct {
	statusConfig *config.StatusNames
}

func NewStatusConverter(statusConfig *config.StatusNames) StatusConverter {
	return StatusConverter{statusConfig}
}

func (sc StatusConverter) Status(originStatus string) Status {
	candidates := map[Status][]string{
		StatusDone:       sc.statusConfig.Done,
		StatusInProgress: sc.statusConfig.InProgress,
		StatusToDo:       sc.statusConfig.ToDo,
	}

	for result, candidateItems := range candidates {
		for _, item := range candidateItems {
			if item == originStatus {
				return result
			}
		}
	}

	return StatusUndefined
}

func (sc StatusConverter) PlanningStatus(
	actualStatus Status,
	snapshotDate time.Time,
	actualStartDate time.Time,
	actualDueDate time.Time,
	historicDueDates []historicDueDate,
) PlanningStatus {
	fmt.Fprintln(os.Stderr, "> historicDueDates:", historicDueDates)

	if len(historicDueDates) > 0 {
		// DESC
		sort.Slice(historicDueDates, func(i, j int) bool { return historicDueDates[i].DueDate.Unix() > historicDueDates[j].DueDate.Unix() })

		if historicDueDates[0].DueDate.Before(actualDueDate) {
			return PlanningStatusPostponed
		}

		if historicDueDates[0].DueDate.After(actualDueDate) {
			return PlanningStatusAdvanced
		}
	}

	pastDueDate := snapshotDate.After(actualDueDate)
	if (!actualStatus.isDone() && pastDueDate) || (actualStatus.isToDo() && snapshotDate.After(actualStartDate)) {
		return PlanningStatusOverdue
	}

	return PlanningStatusOK

}

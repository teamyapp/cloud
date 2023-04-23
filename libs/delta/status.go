package delta

type Status string

const (
	UnchangedStatus Status = "unchanged"
	AddedStatus     Status = "added"
	UpdatedStatus   Status = "updated"
	RemovedStatus   Status = "removed"
)

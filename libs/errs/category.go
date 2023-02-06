package errs

type Category string

const (
	Transient         Category = "Transient"
	Outage            Category = "Outage"
	ClientInteraction Category = "ClientInteraction"
)

package errs

type Category string

const (
	Transient         Category = "Transient"
	Outage            Category = "Outage"
	ClientInteraction Category = "ClientInteraction"
)

var errorCategory = map[ErrorCode]Category{
	NotFound:          ClientInteraction,
	AlreadyExists:     ClientInteraction,
	Unauthenticated:   ClientInteraction,
	PermissionDenied:  ClientInteraction,
	InvalidArgument:   ClientInteraction,
	InvalidValue:      ClientInteraction,
	InvalidFormat:     ClientInteraction,
	InvalidOperation:  ClientInteraction,
	Serialization:     ClientInteraction,
	Deserialization:   ClientInteraction,
	Cancelled:         Transient,
	Aborted:           Transient,
	Timeout:           Transient,
	ResourceExhausted: Transient,
	NotReady:          Transient,
	IO:                Transient,
	Unreachable:       Outage,
	OS:                Outage,
	Unimplemented:     Outage,
	Unknown:           Outage,
}

func GetErrorCategory(errCode ErrorCode) Category {
	category, ok := errorCategory[errCode]
	if ok {
		return category
	}

	return Outage
}

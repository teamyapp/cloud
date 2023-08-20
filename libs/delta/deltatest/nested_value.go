package deltatest

import (
	"github.com/teamyapp/cloud/libs/delta"
)

type NestedValue struct {
	FirstName string
	LastName  string
	Age       int
	IsAdmin   bool
}

type NestedValueDelta struct {
	FirstName delta.Delta[string]
	LastName  delta.Delta[string]
	Age       delta.Delta[int]
	IsAdmin   delta.Delta[bool]
}

func DetectNestedValueDelta(oldValue NestedValue, newValue NestedValue) delta.Delta[NestedValueDelta] {
	status := delta.UnchangedStatus
	firstName := delta.DetectValueDelta(oldValue.FirstName, newValue.FirstName)
	lastName := delta.DetectValueDelta(oldValue.LastName, newValue.LastName)
	age := delta.DetectValueDelta(oldValue.Age, newValue.Age)
	isAdmin := delta.DetectValueDelta(oldValue.IsAdmin, newValue.IsAdmin)

	if firstName.Status != delta.UnchangedStatus ||
		lastName.Status != delta.UnchangedStatus ||
		age.Status != delta.UnchangedStatus ||
		isAdmin.Status != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[NestedValueDelta]{
		Status: status,
		Value: NestedValueDelta{
			FirstName: firstName,
			LastName:  lastName,
			Age:       age,
			IsAdmin:   isAdmin,
		},
	}
}

func ToNestValueDelta(status delta.Status, value NestedValue) NestedValueDelta {
	return NestedValueDelta{
		FirstName: delta.Delta[string]{Status: status, Value: value.FirstName},
		LastName:  delta.Delta[string]{Status: status, Value: value.LastName},
		Age:       delta.Delta[int]{Status: status, Value: value.Age},
		IsAdmin:   delta.Delta[bool]{Status: status, Value: value.IsAdmin},
	}
}

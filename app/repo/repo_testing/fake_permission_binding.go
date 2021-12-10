package repo_testing

import (
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/repo"
)

type FakePermissionBinding struct {
	entities map[entity.PermissionBinding]bool
}

var _ repo.PermissionBinding = (*FakePermissionBinding)(nil)

func (f *FakePermissionBinding) HasPermissionBinding(query entity.PermissionBinding) bool {
	return f.entities[query]
}

func (f *FakePermissionBinding) AddPermissionBinding(binding entity.PermissionBinding) {
	if f.HasPermissionBinding(binding) {
		return
	}
	f.entities[binding] = true
}

func NewFakePermissionBinding() FakePermissionBinding {
	return FakePermissionBinding{entities: make(map[entity.PermissionBinding]bool, 0)}
}

package repo_testing

import (
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/repo"
)

type FakePermissionGraph struct {
	nodeToNeighboursMap map[entity.Permission]map[entity.Permission]bool
}

var _ repo.PermissionGraph = (*FakePermissionGraph)(nil)

func (f *FakePermissionGraph) GetNeighbours(permission entity.Permission) []entity.Permission {
	neighbours, ok := f.nodeToNeighboursMap[permission]
	if !ok {
		return []entity.Permission{}
	}

	response := make([]entity.Permission, 0)
	for neighbour, exists := range neighbours {
		if !exists {
			continue
		}
		response = append(response, neighbour)
	}
	return response
}

func (f *FakePermissionGraph) AddNeighbour(node entity.Permission, neighbour entity.Permission) {
	if f.hasNeighbour(node, neighbour) {
		return
	}

	if _, ok := f.nodeToNeighboursMap[node]; !ok {
		f.nodeToNeighboursMap[node] = make(map[entity.Permission]bool, 0)
	}

	f.nodeToNeighboursMap[node][neighbour] = true
}

func (f *FakePermissionGraph) hasNeighbour(node entity.Permission, neighbour entity.Permission) bool {
	neighbours, ok := f.nodeToNeighboursMap[node]
	if !ok {
		return false
	}

	return neighbours[neighbour]
}

func NewFakePermissionGraph() FakePermissionGraph {
	return FakePermissionGraph{nodeToNeighboursMap: make(map[entity.Permission]map[entity.Permission]bool, 0)}
}

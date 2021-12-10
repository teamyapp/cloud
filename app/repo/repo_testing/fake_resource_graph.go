package repo_testing

import (
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/repo"
)

type FakeResourceGraph struct {
	nodeToNeighboursMap map[entity.Resource]map[entity.Resource]bool
}

var _ repo.ResourceGraph = (*FakeResourceGraph)(nil)

func (f *FakeResourceGraph) FindNeighboursWithType(resource entity.Resource, resourceType string) []entity.Resource {
	neighbours, ok := f.nodeToNeighboursMap[resource]
	if !ok {
		return []entity.Resource{}
	}

	response := make([]entity.Resource, 0)
	for neighbour, exists := range neighbours {
		if !exists {
			continue
		}
		if neighbour.Type != resourceType {
			continue
		}
		response = append(response, neighbour)
	}
	return response
}

func (f *FakeResourceGraph) AddNeighbour(node entity.Resource, neighbour entity.Resource) {
	if f.hasNeighbour(node, neighbour) {
		return
	}

	if _, ok := f.nodeToNeighboursMap[node]; !ok {
		f.nodeToNeighboursMap[node] = make(map[entity.Resource]bool, 0)
	}

	f.nodeToNeighboursMap[node][neighbour] = true
}

func (f *FakeResourceGraph) hasNeighbour(node entity.Resource, neighbour entity.Resource) bool {
	neighbours, ok := f.nodeToNeighboursMap[node]
	if !ok {
		return false
	}

	return neighbours[neighbour]
}

func NewFakeResourceGraph() FakeResourceGraph {
	return FakeResourceGraph{nodeToNeighboursMap: make(map[entity.Resource]map[entity.Resource]bool, 0)}
}

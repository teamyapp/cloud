package repo

import "github.com/teamyapp/cloud/app/entity"

type ResourceGraph interface {
	FindNeighboursWithType(resource entity.Resource, resourceType string) []entity.Resource
	AddNeighbour(node entity.Resource, neighbour entity.Resource)
}

package entity

import "context"

type Iterator[Item any] interface {
	HasNext() (bool, error)
	Next(ct context.Context) (Item, error)
}

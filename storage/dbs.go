package storage

import (
	"errors"
)

type HodBucket string

var (
	EntityBucket   HodBucket = "entity"
	PKBucket       HodBucket = "pk"
	PredBucket     HodBucket = "pred"
	GraphBucket    HodBucket = "graph"
	ExtendedBucket HodBucket = "extended"
)

var (
	ErrNotFound = errors.New("Not found")
)

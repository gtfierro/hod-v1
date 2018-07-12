package storage

import (
	"errors"
)

type HodNamespace string

var (
	EntityBucket   HodNamespace = "entity"
	PKBucket       HodNamespace = "pk"
	PredBucket     HodNamespace = "pred"
	GraphBucket    HodNamespace = "graph"
	ExtendedBucket HodNamespace = "extended"
)

var (
	ErrNotFound = errors.New("Not found")
)

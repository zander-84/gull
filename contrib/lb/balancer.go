package lb

import (
	"errors"
)

// ErrNoNode is returned when no qualifying node are available.
var ErrNoNode = errors.New("no node available")
var ErrNotImplemented = errors.New("not implemented")

type Listener interface {
	CleanErr()
	RecordErr(err error)
	Err() error
	Get() (map[any]int, []any, uint64)
	GetVersion() uint64
	Remove(addr any) error
	AddWeight(addr any, weight int) error
	Add(addr any) error
	Set(data map[any]int) error
	Exist(addr any) bool
	Close()
}

// Balancer yields endpoints according to some heuristic.
type Balancer interface {
	Update()
	Next() (any, error)
	All() ([]any, error)
	Used() map[any]int64
	Get(uid any) (any, error) //用于hash一致性
}

type Policy int

const (
	RoundRobin Policy = iota
	WeightRoundRobin
	ConsistentHash
	Random
)

func NewBalancer(listener Listener, p Policy, record bool) Balancer {
	switch p {
	case RoundRobin:
		return NewRoundRobin(listener, record)
	case WeightRoundRobin:
		return NewWeightRoundRobin(listener, record)
	case ConsistentHash:
		return NewConsistentHash(listener, record)
	case Random:
		return NewRandom(listener, record)
	default:
		return NewRoundRobin(listener, record)
	}
}

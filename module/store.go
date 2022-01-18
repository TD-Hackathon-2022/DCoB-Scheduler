package module

import (
	"sync"
)

type JobStore interface {
	Store(job Job)
	Load(jobId string) (job Job, exist bool)
}

type simpleStore struct {
	innerMap sync.Map
}

func (s *simpleStore) Store(job Job) {
	s.innerMap.Store(job.Id(), job)
}

func (s *simpleStore) Load(jobId string) (job Job, exist bool) {
	v, exist := s.innerMap.Load(jobId)
	if !exist {
		return nil, false
	}

	return v.(Job), true
}

func NewSimpleStore() JobStore {
	return &simpleStore{}
}

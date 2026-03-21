package runtime

import "ctf-platform/internal/module/container"

type RuntimeQuery interface {
	CountRunning() (int64, error)
}

type Query struct {
	repo *container.Repository
}

func NewQuery(repo *container.Repository) *Query {
	return &Query{repo: repo}
}

func (q *Query) CountRunning() (int64, error) {
	if q == nil || q.repo == nil {
		return 0, nil
	}
	return q.repo.CountRunning()
}

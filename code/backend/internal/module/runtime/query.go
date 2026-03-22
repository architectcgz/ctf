package runtime

type RuntimeQuery interface {
	CountRunning() (int64, error)
}

type Query struct {
	repo *Repository
}

func NewQuery(repo *Repository) *Query {
	return &Query{repo: repo}
}

func (q *Query) CountRunning() (int64, error) {
	if q == nil || q.repo == nil {
		return 0, nil
	}
	return q.repo.CountRunning()
}

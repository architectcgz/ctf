package ports

import "context"

type RuntimeNodeBinding struct {
	NodeID   int64
	NodeName string
}

type RuntimeNodeSelector interface {
	SelectDefaultNode(ctx context.Context) (*RuntimeNodeBinding, error)
}

type RuntimeNodeBootstrapSpec struct {
	Name        string
	Endpoint    string
	TLSIdentity string
	Schedulable bool
}

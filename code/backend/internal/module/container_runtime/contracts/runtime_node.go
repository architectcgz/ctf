package contracts

type RuntimeNodeBinding struct {
	NodeID   int64
	NodeName string
}

type RuntimeNodeBootstrapSpec struct {
	Name        string
	Endpoint    string
	PublicHost  string
	AccessHost  string
	TLSIdentity string
	Schedulable bool
}

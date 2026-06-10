package application

import "testing"

func TestTopologyContractsCompile(t *testing.T) {
	_ = TopologyCreateRequest{}
	_ = TopologyCreateResult{}
	_ = TopologyCreateNode{}
	_ = TopologyCreateNetwork{}
}

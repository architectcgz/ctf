package model

import "testing"

func TestResolveRuntimeAliasAccessURLUsesEntryNetworkIP(t *testing.T) {
	t.Parallel()

	details := `{
		"networks":[
			{"key":"default","name":"ctf-awd-contest-8","network_id":"net-awd-contest-8","shared":true}
		],
		"containers":[
			{
				"container_id":"ctr-awd",
				"is_entry_point":true,
				"network_keys":["default"],
				"network_aliases":["awd-c8-t15-s21"],
				"network_ips":{"ctf-awd-contest-8":"172.30.0.20"}
			}
		]
	}`

	got := ResolveRuntimeAliasAccessURL("http://awd-c8-t15-s21:8080", details)
	if got != "http://172.30.0.20:8080" {
		t.Fatalf("unexpected resolved access url: %s", got)
	}
}

func TestResolveRuntimeAliasAccessURLKeepsStableURLWhenIPMissing(t *testing.T) {
	t.Parallel()

	got := ResolveRuntimeAliasAccessURL("http://awd-c8-t15-s21:8080", `{"containers":[{"is_entry_point":true}]}`)
	if got != "http://awd-c8-t15-s21:8080" {
		t.Fatalf("expected stable alias url to remain unchanged, got %s", got)
	}
}

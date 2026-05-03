package evidence

import "testing"

func TestBuildProxyRequestDetailAndMeta(t *testing.T) {
	raw := `{"method":"post","target_path":"/login","target_query":"a=1","status":200,"payload_preview":"username=admin"}`

	detail := BuildProxyRequestDetail(raw)
	if detail != "经平台代理发起 POST /login?a=1，请求返回 200，携带请求摘要" {
		t.Fatalf("unexpected detail: %s", detail)
	}

	meta := BuildProxyRequestMeta(raw)
	if meta["request_method"] != "POST" {
		t.Fatalf("unexpected request_method: %+v", meta)
	}
	if meta["target_path"] != "/login" {
		t.Fatalf("unexpected target_path: %+v", meta)
	}
	if meta["target_query"] != "a=1" {
		t.Fatalf("unexpected target_query: %+v", meta)
	}
	if meta["status_code"] != 200 {
		t.Fatalf("unexpected status_code: %+v", meta)
	}
	if meta["payload_preview"] != "username=admin" {
		t.Fatalf("unexpected payload_preview: %+v", meta)
	}
}

func TestBuildSharedEvidenceDetails(t *testing.T) {
	if BuildManualReviewDetail("approved") != "人工评审已通过" {
		t.Fatal("expected approved manual review detail")
	}
	if BuildAWDAttackDetail(true) != "AWD 攻击提交成功" {
		t.Fatal("expected success awd attack detail")
	}
	if BuildAWDTrafficDetail("post", "/flag", 403) != "AWD 流量 POST /flag 返回 403" {
		t.Fatal("expected awd traffic detail")
	}
}

package application

import (
	"testing"
	"time"
)

func TestAggregateSubmitBurstsAppliesThresholdSortAndLimit(t *testing.T) {
	base := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
	events := make([]RiskAuditEvent, 0)

	appendEvents := func(userID int64, username string, count int, lastSeen time.Time) {
		for idx := 0; idx < count; idx++ {
			eventTime := lastSeen.Add(-time.Duration(count-idx-1) * time.Second)
			events = append(events, RiskAuditEvent{
				UserID:    &userID,
				Username:  username,
				IPAddress: "10.0.0.1",
				CreatedAt: eventTime,
			})
		}
	}

	appendEvents(1, "u01", submitBurstMinCount, base.Add(1*time.Minute))
	appendEvents(2, "u02", submitBurstMinCount, base.Add(2*time.Minute))
	for userID := int64(3); userID <= 11; userID++ {
		appendEvents(userID, "u", int(userID)+3, base.Add(time.Duration(userID)*time.Minute))
	}
	appendEvents(99, "ignored", submitBurstMinCount-1, base.Add(99*time.Minute))

	rows, userIDs := aggregateSubmitBursts(events)

	if len(rows) != maxRiskRows {
		t.Fatalf("aggregateSubmitBursts() rows len = %d, want %d", len(rows), maxRiskRows)
	}
	if len(userIDs) != 11 {
		t.Fatalf("aggregateSubmitBursts() qualifying userIDs len = %d, want 11", len(userIDs))
	}
	if _, ok := userIDs[99]; ok {
		t.Fatal("aggregateSubmitBursts() should exclude below-threshold user from qualifying set")
	}
	if rows[0].UserID != 11 || rows[0].SubmitCount != 14 {
		t.Fatalf("aggregateSubmitBursts() top row = %+v, want user 11 with count 14", rows[0])
	}
	if rows[len(rows)-1].UserID != 2 || rows[len(rows)-1].SubmitCount != submitBurstMinCount {
		t.Fatalf("aggregateSubmitBursts() last row = %+v, want tie-broken user 2", rows[len(rows)-1])
	}
	for _, row := range rows {
		if row.UserID == 1 {
			t.Fatal("aggregateSubmitBursts() should drop older tie when max rows exceeded")
		}
	}
}

func TestAggregateSharedIPsGroupsUsersAndSorts(t *testing.T) {
	now := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
	user1 := int64(1)
	user2 := int64(2)
	user3 := int64(3)
	user4 := int64(4)
	user5 := int64(5)
	user6 := int64(6)
	user7 := int64(7)
	user8 := int64(8)

	rows, affectedUsers := aggregateSharedIPs([]RiskAuditEvent{
		{UserID: &user2, Username: "bravo", IPAddress: "10.0.0.1", CreatedAt: now},
		{UserID: &user1, Username: "alpha", IPAddress: "10.0.0.1", CreatedAt: now},
		{UserID: &user1, Username: "alpha", IPAddress: "10.0.0.1", CreatedAt: now},
		{UserID: &user3, Username: "charlie", IPAddress: "10.0.0.2", CreatedAt: now},
		{UserID: &user4, Username: "delta", IPAddress: "10.0.0.2", CreatedAt: now},
		{UserID: &user5, Username: "echo", IPAddress: "10.0.0.2", CreatedAt: now},
		{UserID: &user6, Username: "foxtrot", IPAddress: "10.0.0.0", CreatedAt: now},
		{UserID: &user7, Username: "golf", IPAddress: "10.0.0.0", CreatedAt: now},
		{UserID: &user8, Username: "hotel", IPAddress: "10.0.0.0", CreatedAt: now},
		{UserID: nil, Username: "ignored", IPAddress: "10.0.0.9", CreatedAt: now},
		{UserID: &user8, Username: "hotel", IPAddress: "", CreatedAt: now},
	})

	if len(rows) != 3 {
		t.Fatalf("aggregateSharedIPs() rows len = %d, want 3", len(rows))
	}
	if len(affectedUsers) != 8 {
		t.Fatalf("aggregateSharedIPs() affectedUsers len = %d, want 8", len(affectedUsers))
	}
	if rows[0].IP != "10.0.0.0" || rows[0].UserCount != 3 {
		t.Fatalf("aggregateSharedIPs() first row = %+v, want 10.0.0.0 with 3 users", rows[0])
	}
	if rows[1].IP != "10.0.0.2" || rows[1].UserCount != 3 {
		t.Fatalf("aggregateSharedIPs() second row = %+v, want 10.0.0.2 with 3 users", rows[1])
	}
	if rows[2].IP != "10.0.0.1" || rows[2].UserCount != 2 {
		t.Fatalf("aggregateSharedIPs() third row = %+v, want 10.0.0.1 with 2 users", rows[2])
	}
	if len(rows[2].Usernames) != 2 || rows[2].Usernames[0] != "alpha" || rows[2].Usernames[1] != "bravo" {
		t.Fatalf("aggregateSharedIPs() usernames = %+v, want sorted unique usernames", rows[2].Usernames)
	}
}

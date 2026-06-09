package commands

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

func TestPackageDeliveryServiceDelegatesJeopardyPreviewAndCommit(t *testing.T) {
	jeopardy := &stubJeopardyPackageDelivery{
		preview: &challengecontracts.ChallengeImportPreviewResp{
			ID:        "jeopardy-preview",
			Title:     "Jeopardy Preview",
			CreatedAt: time.Date(2026, 6, 9, 1, 0, 0, 0, time.UTC),
		},
		commit: &challengecontracts.ChallengeResp{ID: 42, Title: "Imported"},
	}
	service := NewPackageDeliveryService(jeopardy, nil)

	preview, err := service.Preview(context.Background(), PackageDeliveryPreviewRequest{
		Mode:        PackageDeliveryModeJeopardy,
		ActorUserID: 1001,
		FileName:    "challenge.zip",
		Reader:      strings.NewReader("zip"),
	})
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if preview.Jeopardy == nil || preview.Jeopardy.ID != "jeopardy-preview" {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	if !jeopardy.previewCalled {
		t.Fatal("expected Jeopardy preview delegate")
	}

	commit, err := service.Commit(context.Background(), PackageDeliveryCommitRequest{
		Mode:        PackageDeliveryModeJeopardy,
		ActorUserID: 1001,
		PreviewID:   "jeopardy-preview",
	})
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if commit.Jeopardy == nil || commit.Jeopardy.ID != 42 {
		t.Fatalf("unexpected commit: %+v", commit)
	}
	if !jeopardy.commitCalled {
		t.Fatal("expected Jeopardy commit delegate")
	}
}

func TestPackageDeliveryServiceDelegatesAWDPreviewAndCommit(t *testing.T) {
	awd := &stubAWDPackageDelivery{
		preview: &challengecontracts.AWDChallengeImportPreviewResp{
			ID:        "awd-preview",
			Title:     "AWD Preview",
			CreatedAt: time.Date(2026, 6, 9, 1, 0, 0, 0, time.UTC),
		},
		commit: &challengecontracts.AWDChallengeResp{ID: 77, Name: "AWD Imported"},
	}
	service := NewPackageDeliveryService(nil, awd)

	preview, err := service.Preview(context.Background(), PackageDeliveryPreviewRequest{
		Mode:        PackageDeliveryModeAWD,
		ActorUserID: 1001,
		FileName:    "awd.zip",
		Reader:      strings.NewReader("zip"),
	})
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if preview.AWD == nil || preview.AWD.ID != "awd-preview" {
		t.Fatalf("unexpected preview: %+v", preview)
	}

	commit, err := service.Commit(context.Background(), PackageDeliveryCommitRequest{
		Mode:        PackageDeliveryModeAWD,
		ActorUserID: 1001,
		PreviewID:   "awd-preview",
	})
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if commit.AWD == nil || commit.AWD.ID != 77 {
		t.Fatalf("unexpected commit: %+v", commit)
	}
	if !awd.previewCalled || !awd.commitCalled {
		t.Fatalf("expected AWD delegates, got preview=%v commit=%v", awd.previewCalled, awd.commitCalled)
	}
}

type stubJeopardyPackageDelivery struct {
	preview       *challengecontracts.ChallengeImportPreviewResp
	commit        *challengecontracts.ChallengeResp
	previewCalled bool
	commitCalled  bool
}

func (s *stubJeopardyPackageDelivery) PreviewChallengeImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.ChallengeImportPreviewResp, error) {
	s.previewCalled = true
	return s.preview, nil
}

func (s *stubJeopardyPackageDelivery) CommitChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeResp, error) {
	s.commitCalled = true
	return s.commit, nil
}

type stubAWDPackageDelivery struct {
	preview       *challengecontracts.AWDChallengeImportPreviewResp
	commit        *challengecontracts.AWDChallengeResp
	previewCalled bool
	commitCalled  bool
}

func (s *stubAWDPackageDelivery) PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.AWDChallengeImportPreviewResp, error) {
	s.previewCalled = true
	return s.preview, nil
}

func (s *stubAWDPackageDelivery) CommitImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeResp, error) {
	s.commitCalled = true
	return s.commit, nil
}

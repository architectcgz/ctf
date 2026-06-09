package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

const (
	PackageDeliveryModeJeopardy = "jeopardy"
	PackageDeliveryModeAWD      = "awd"
)

type packageDeliveryJeopardyService interface {
	PreviewChallengeImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.ChallengeImportPreviewResp, error)
	CommitChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeResp, error)
}

type packageDeliveryAWDService interface {
	PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.AWDChallengeImportPreviewResp, error)
	CommitImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeResp, error)
}

type PackageDeliveryService struct {
	jeopardy packageDeliveryJeopardyService
	awd      packageDeliveryAWDService
}

type PackageDeliveryPreviewRequest struct {
	Mode        string
	ActorUserID int64
	FileName    string
	Reader      io.Reader
}

type PackageDeliveryPreviewResult struct {
	Mode     string
	Jeopardy *challengecontracts.ChallengeImportPreviewResp
	AWD      *challengecontracts.AWDChallengeImportPreviewResp
}

type PackageDeliveryCommitRequest struct {
	Mode        string
	ActorUserID int64
	PreviewID   string
}

type PackageDeliveryCommitResult struct {
	Mode     string
	Jeopardy *challengecontracts.ChallengeResp
	AWD      *challengecontracts.AWDChallengeResp
}

func NewPackageDeliveryService(jeopardy packageDeliveryJeopardyService, awd packageDeliveryAWDService) *PackageDeliveryService {
	return &PackageDeliveryService{jeopardy: jeopardy, awd: awd}
}

func (s *PackageDeliveryService) Preview(ctx context.Context, req PackageDeliveryPreviewRequest) (*PackageDeliveryPreviewResult, error) {
	mode := normalizePackageDeliveryMode(req.Mode)
	switch mode {
	case PackageDeliveryModeJeopardy:
		if s == nil || s.jeopardy == nil {
			return nil, fmt.Errorf("jeopardy package delivery service is not configured")
		}
		preview, err := s.jeopardy.PreviewChallengeImport(ctx, req.ActorUserID, req.FileName, req.Reader)
		if err != nil {
			return nil, err
		}
		return &PackageDeliveryPreviewResult{Mode: mode, Jeopardy: preview}, nil
	case PackageDeliveryModeAWD:
		if s == nil || s.awd == nil {
			return nil, fmt.Errorf("awd package delivery service is not configured")
		}
		preview, err := s.awd.PreviewImport(ctx, req.ActorUserID, req.FileName, req.Reader)
		if err != nil {
			return nil, err
		}
		return &PackageDeliveryPreviewResult{Mode: mode, AWD: preview}, nil
	default:
		return nil, fmt.Errorf("unsupported package delivery mode %q", req.Mode)
	}
}

func (s *PackageDeliveryService) Commit(ctx context.Context, req PackageDeliveryCommitRequest) (*PackageDeliveryCommitResult, error) {
	mode := normalizePackageDeliveryMode(req.Mode)
	switch mode {
	case PackageDeliveryModeJeopardy:
		if s == nil || s.jeopardy == nil {
			return nil, fmt.Errorf("jeopardy package delivery service is not configured")
		}
		challenge, err := s.jeopardy.CommitChallengeImport(ctx, req.ActorUserID, req.PreviewID)
		if err != nil {
			return nil, err
		}
		return &PackageDeliveryCommitResult{Mode: mode, Jeopardy: challenge}, nil
	case PackageDeliveryModeAWD:
		if s == nil || s.awd == nil {
			return nil, fmt.Errorf("awd package delivery service is not configured")
		}
		challenge, err := s.awd.CommitImport(ctx, req.ActorUserID, req.PreviewID)
		if err != nil {
			return nil, err
		}
		return &PackageDeliveryCommitResult{Mode: mode, AWD: challenge}, nil
	default:
		return nil, fmt.Errorf("unsupported package delivery mode %q", req.Mode)
	}
}

func normalizePackageDeliveryMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

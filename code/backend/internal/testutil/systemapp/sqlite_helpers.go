package systemapp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	containerruntimeentity "ctf-platform/internal/module/container_runtime/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	practiceentity "ctf-platform/internal/module/practice/entity"
	"ctf-platform/internal/platform/clustersecret"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type appImageRow struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	Name        string         `gorm:"column:name;uniqueIndex:idx_name_tag"`
	Tag         string         `gorm:"column:tag;uniqueIndex:idx_name_tag"`
	Description string         `gorm:"column:description"`
	Size        int64          `gorm:"column:size"`
	Status      string         `gorm:"column:status"`
	Digest      string         `gorm:"column:digest"`
	SourceType  string         `gorm:"column:source_type"`
	BuildJobID  *int64         `gorm:"column:build_job_id"`
	LastError   string         `gorm:"column:last_error"`
	VerifiedAt  *time.Time     `gorm:"column:verified_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (appImageRow) TableName() string {
	return "images"
}

type appChallengeRow struct {
	ID              int64          `gorm:"column:id;primaryKey"`
	PackageSlug     *string        `gorm:"column:package_slug;size:128;uniqueIndex:uq_challenges_package_slug"`
	Title           string         `gorm:"column:title"`
	Description     string         `gorm:"column:description"`
	Category        string         `gorm:"column:category"`
	Difficulty      string         `gorm:"column:difficulty"`
	Points          int            `gorm:"column:points"`
	ImageID         int64          `gorm:"column:image_id"`
	AttachmentURL   string         `gorm:"column:attachment_url"`
	Status          string         `gorm:"column:status"`
	FlagType        string         `gorm:"column:flag_type;default:'static'"`
	FlagHash        string         `gorm:"column:flag_hash;size:128"`
	FlagSalt        string         `gorm:"column:flag_salt;size:64"`
	FlagRegex       string         `gorm:"column:flag_regex;size:512"`
	FlagPrefix      string         `gorm:"column:flag_prefix;size:32;default:'flag'"`
	InstanceSharing string         `gorm:"column:instance_sharing;size:16;default:'per_user'"`
	TargetProtocol  string         `gorm:"column:target_protocol;size:16;default:'http'"`
	TargetPort      int            `gorm:"column:target_port;default:0"`
	CreatedBy       *int64         `gorm:"column:created_by"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (appChallengeRow) TableName() string {
	return "challenges"
}

var internalAppTestSchemaModels = []any{
	&identitycontracts.Role{},
	&identitycontracts.User{},
	&identitycontracts.UserRole{},
	&appImageRow{},
	&appChallengeRow{},
	&challengecontracts.AWDChallenge{},
	&challengeentity.ChallengePublishCheckJob{},
	&challengeentity.Tag{},
	&challengeentity.ChallengeTag{},
	&challengeentity.ChallengeHint{},
	&challengeentity.ChallengeWriteup{},
	&challengeentity.SubmissionWriteup{},
	&challengeentity.EnvironmentTemplate{},
	&challengeentity.ChallengeTopology{},
	&challengeentity.ChallengePackageRevision{},
	&contestcontracts.Submission{},
	&instancecontracts.Instance{},
	&clustersecret.RuntimeClusterSecret{},
	&containerruntimeentity.PortAllocation{},
	&containerruntimeentity.NetworkAllocation{},
	&containerruntimeentity.RuntimeNode{},
	&containerruntimeentity.RuntimePortPool{},
	&containerruntimeentity.RuntimeSubnetPool{},
	&practiceentity.UserScore{},
	&opsentity.AuditLog{},
	&opsentity.NotificationBatch{},
	&opsentity.Notification{},
	&assessmententity.SkillProfile{},
	&assessmententity.Report{},
	&contestcontracts.Contest{},
	&contestentity.ContestStatusTransition{},
	&contestcontracts.ContestChallenge{},
	&contestcontracts.ContestAWDService{},
	&contestcontracts.ContestRegistration{},
	&contestentity.ContestAnnouncement{},
	&contestcontracts.Team{},
	&contestcontracts.TeamMember{},
	&contestcontracts.AWDRound{},
	&contestcontracts.AWDTeamService{},
	&contestcontracts.AWDAttackLog{},
	&contestcontracts.AWDTrafficEvent{},
	&contestcontracts.AWDServiceOperation{},
	&contestcontracts.AWDScopeControl{},
	&contestcontracts.AWDDefenseWorkspace{},
}

var (
	internalAppTestSchemaTemplateOnce sync.Once
	internalAppTestSchemaTemplatePath string
	internalAppTestSchemaTemplateErr  error
	internalAppTestPortRangeMu        sync.Mutex
	internalAppTestPortRangeReserved  = make(map[int]int)
)

func OpenInternalAppTestSQLite(t *testing.T, filename string) *gorm.DB {
	t.Helper()

	templatePath, err := ensureInternalAppTestSchemaTemplate()
	if err != nil {
		t.Fatalf("prepare internal app schema template: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), filename)
	if err := copyInternalAppSQLiteTemplate(templatePath, dbPath); err != nil {
		t.Fatalf("clone internal app sqlite template: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func ensureInternalAppTestSchemaTemplate() (string, error) {
	internalAppTestSchemaTemplateOnce.Do(func() {
		dir, err := os.MkdirTemp("", "internal-app-schema-*")
		if err != nil {
			internalAppTestSchemaTemplateErr = fmt.Errorf("create schema temp dir: %w", err)
			return
		}

		internalAppTestSchemaTemplatePath = filepath.Join(dir, "schema.sqlite")
		internalAppTestSchemaTemplateErr = buildInternalAppTestSchemaTemplate(internalAppTestSchemaTemplatePath)
	})

	if internalAppTestSchemaTemplateErr != nil {
		return "", internalAppTestSchemaTemplateErr
	}
	return internalAppTestSchemaTemplatePath, nil
}

func buildInternalAppTestSchemaTemplate(path string) error {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open sqlite template: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sqlite template handle: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.AutoMigrate(internalAppTestSchemaModels...); err != nil {
		return fmt.Errorf("auto migrate sqlite template: %w", err)
	}

	return nil
}

func copyInternalAppSQLiteTemplate(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open sqlite template: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create sqlite copy: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return fmt.Errorf("copy sqlite template: %w", err)
	}
	if err := dst.Sync(); err != nil {
		_ = dst.Close()
		return fmt.Errorf("sync sqlite copy: %w", err)
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("close sqlite copy: %w", err)
	}
	return nil
}

func ReserveInternalAppTestPortRange(t *testing.T, width int) (int, int) {
	t.Helper()

	if width <= 0 {
		t.Fatalf("invalid test port range width: %d", width)
	}

	const (
		candidateStart = 30000
		candidateLimit = 61000
	)

	internalAppTestPortRangeMu.Lock()
	defer internalAppTestPortRangeMu.Unlock()

	for start := candidateStart; start+width-1 <= candidateLimit; start += width {
		end := start + width - 1
		if internalAppTestPortRangeOverlaps(start, end) {
			continue
		}
		if !internalAppTestPortRangeAvailable(start, end) {
			continue
		}

		internalAppTestPortRangeReserved[start] = end
		t.Cleanup(func() {
			internalAppTestPortRangeMu.Lock()
			delete(internalAppTestPortRangeReserved, start)
			internalAppTestPortRangeMu.Unlock()
		})
		return start, end
	}

	t.Fatalf("no available internal app test port range with width %d", width)
	return 0, 0
}

func internalAppTestPortRangeOverlaps(start, end int) bool {
	for reservedStart, reservedEnd := range internalAppTestPortRangeReserved {
		if end < reservedStart || start > reservedEnd {
			continue
		}
		return true
	}
	return false
}

func internalAppTestPortRangeAvailable(start, end int) bool {
	listeners := make([]net.Listener, 0, end-start+1)
	for port := start; port <= end; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			for _, item := range listeners {
				_ = item.Close()
			}
			return false
		}
		listeners = append(listeners, listener)
	}
	for _, item := range listeners {
		_ = item.Close()
	}
	return true
}

func OccupyTestPortIfFree(t *testing.T, port int) {
	t.Helper()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err == nil {
		t.Cleanup(func() { _ = listener.Close() })
		return
	}
	if !strings.Contains(strings.ToLower(err.Error()), "address already in use") {
		t.Fatalf("occupy port %d: %v", port, err)
	}
}

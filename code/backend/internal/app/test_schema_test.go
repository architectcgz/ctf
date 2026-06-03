package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	practiceentity "ctf-platform/internal/module/practice/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	&runtimeentity.PortAllocation{},
	&runtimeentity.NetworkAllocation{},
	&runtimeentity.RuntimeNode{},
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
	&runtimecontracts.AWDServiceOperation{},
	&runtimecontracts.AWDScopeControl{},
}

var (
	internalAppTestSchemaTemplateOnce sync.Once
	internalAppTestSchemaTemplatePath string
	internalAppTestSchemaTemplateErr  error
)

func openInternalAppTestSQLite(t *testing.T, filename string) *gorm.DB {
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

func TestInternalAppTestSchemaIncludesRuntimeTables(t *testing.T) {
	db := openInternalAppTestSQLite(t, "schema-check.sqlite")

	for _, table := range []string{"network_allocations", "runtime_nodes"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("expected shared internal app test schema to include table %s", table)
		}
	}
}

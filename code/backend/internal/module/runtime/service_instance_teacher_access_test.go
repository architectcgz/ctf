package runtime_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

func TestServiceListTeacherInstancesScopesTeacherAndAppliesFilters(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &identitycontracts.User{ID: 1, Username: "teacher-a", Role: identitycontracts.RoleTeacher, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: identitycontracts.RoleStudent, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 3, Username: "bob", StudentNo: "S-1002", Role: identitycontracts.RoleStudent, ClassName: "Class B", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &runtimeChallengeTestRow{ID: 11, Title: "web-101", Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 101, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 102, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 103, UserID: 2, ChallengeID: 11, ContainerID: "inst-stopped", Status: instanceentity.InstanceStatusStopped, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})

	pageResp, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{})
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(pageResp.List) != 1 {
		t.Fatalf("expected 1 visible instance, got %d (%+v)", len(pageResp.List), pageResp.List)
	}
	if pageResp.List[0].StudentUsername != "alice" || pageResp.List[0].ClassName != "Class A" {
		t.Fatalf("unexpected item: %+v", pageResp.List[0])
	}

	filtered, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{
		Keyword:   "ali",
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with filters error = %v", err)
	}
	if len(filtered.List) != 1 || filtered.List[0].ID != 101 {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}

	byStudentNoKeyword, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{
		Keyword: "1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with student_no keyword error = %v", err)
	}
	if len(byStudentNoKeyword.List) != 1 || byStudentNoKeyword.List[0].ID != 101 {
		t.Fatalf("expected keyword to match student_no, got %+v", byStudentNoKeyword)
	}
}

func TestServiceListTeacherInstancesRejectsTeacherCrossClassFilter(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &identitycontracts.User{ID: 1, Username: "teacher-a", Role: identitycontracts.RoleTeacher, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})

	_, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{ClassName: "Class B"})
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestRepositoryFindUserByIDIgnoresSoftDeletedUsers(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	user := &identitycontracts.User{
		ID:        1,
		Username:  "teacher-a",
		Role:      identitycontracts.RoleTeacher,
		ClassName: "Class A",
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	seedUser(t, repo.db, user)
	if err := repo.db.Delete(user).Error; err != nil {
		t.Fatalf("soft delete user: %v", err)
	}

	got, err := repo.FindUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v", err)
	}
	if got != nil {
		t.Fatalf("expected soft-deleted user to be hidden, got %+v", got)
	}
}

func TestServiceDestroyTeacherInstanceHonorsClassScope(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &identitycontracts.User{ID: 1, Username: "teacher-a", Role: identitycontracts.RoleTeacher, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 2, Username: "alice", Role: identitycontracts.RoleStudent, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 3, Username: "bob", Role: identitycontracts.RoleStudent, ClassName: "Class B", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &runtimeChallengeTestRow{ID: 11, Title: "web-101", Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 201, UserID: 2, ChallengeID: 11, Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 202, UserID: 3, ChallengeID: 11, Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, identitycontracts.RoleTeacher); err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, identitycontracts.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	instance, err := repo.FindByID(context.Background(), 201)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != instanceentity.InstanceStatusStopping {
		t.Fatalf("expected stopping status, got %s", instance.Status)
	}
}

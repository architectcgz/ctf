package commands

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

type AdminService struct {
	repo adminCommandRepository
	log  *zap.Logger
}

type adminCommandRepository interface {
	identitycontracts.UserLookupRepository
	identitycontracts.UserWriteRepository
}

var _ identitycontracts.AdminCommandService = (*AdminService)(nil)

func NewAdminService(repo adminCommandRepository, log *zap.Logger) *AdminService {
	if log == nil {
		log = zap.NewNop()
	}
	return &AdminService{
		repo: repo,
		log:  log,
	}
}

func (s *AdminService) CreateUser(ctx context.Context, req identitycontracts.CreateUserInput) (*dto.AdminUserResp, error) {
	username := strings.TrimSpace(req.Username)
	if existing, err := s.repo.FindByUsername(ctx, username); err == nil && existing != nil {
		return nil, errcode.ErrUsernameExists
	} else if err != nil && !errors.Is(err, identitycontracts.ErrUserNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	identity := normalizeIdentityNumbers(req.Role, req.StudentNo, req.TeacherNo)
	user := &model.User{
		Username:  username,
		Name:      strings.TrimSpace(req.Name),
		Email:     strings.TrimSpace(req.Email),
		StudentNo: identity.StudentNo,
		TeacherNo: identity.TeacherNo,
		Role:      req.Role,
		ClassName: strings.TrimSpace(req.ClassName),
		Status:    defaultUserStatus(req.Status),
	}
	if err := user.SetPassword(req.Password); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, mapServiceError(err)
	}

	resp := toAdminUserResp(user)
	return &resp, nil
}

func (s *AdminService) UpdateUser(ctx context.Context, userID int64, req identitycontracts.UpdateUserInput) (*dto.AdminUserResp, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, mapServiceError(err)
	}

	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		if err := user.SetPassword(strings.TrimSpace(*req.Password)); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}
	if req.Name != nil {
		user.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		user.Email = strings.TrimSpace(*req.Email)
	}
	if req.StudentNo != nil {
		user.StudentNo = strings.TrimSpace(*req.StudentNo)
	}
	if req.TeacherNo != nil {
		user.TeacherNo = strings.TrimSpace(*req.TeacherNo)
	}
	if req.ClassName != nil {
		user.ClassName = strings.TrimSpace(*req.ClassName)
	}
	if req.Role != nil && strings.TrimSpace(*req.Role) != "" {
		user.Role = strings.TrimSpace(*req.Role)
	}
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		user.Status = strings.TrimSpace(*req.Status)
	}

	identity := normalizeIdentityNumbers(user.Role, user.StudentNo, user.TeacherNo)
	user.StudentNo = identity.StudentNo
	user.TeacherNo = identity.TeacherNo

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, mapServiceError(err)
	}

	resp := toAdminUserResp(user)
	return &resp, nil
}

func (s *AdminService) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.repo.Delete(ctx, userID); err != nil {
		return mapServiceError(err)
	}
	return nil
}

func (s *AdminService) ImportUsers(ctx context.Context, reader io.Reader) (*dto.ImportUsersResp, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	result := &dto.ImportUsersResp{
		Errors: make([]dto.ImportUserError, 0),
	}

	rowIndex := 0
	for {
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errcode.New(errcode.ErrInvalidParams.Code, "导入文件格式错误", errcode.ErrInvalidParams.HTTPStatus)
		}
		rowIndex++
		if rowIndex == 1 && looksLikeHeader(record) {
			continue
		}
		if isBlankRecord(record) {
			continue
		}

		created, err := s.importRow(ctx, record)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, dto.ImportUserError{
				Row:     rowIndex,
				Message: err.Error(),
			})
			continue
		}

		if created {
			result.Created++
		} else {
			result.Updated++
		}
	}

	return result, nil
}

func (s *AdminService) importRow(ctx context.Context, record []string) (bool, error) {
	username := strings.TrimSpace(getCSVValue(record, 0))
	password := strings.TrimSpace(getCSVValue(record, 1))
	email := strings.TrimSpace(getCSVValue(record, 2))
	className := strings.TrimSpace(getCSVValue(record, 3))
	role := strings.TrimSpace(getCSVValue(record, 4))
	status := strings.TrimSpace(getCSVValue(record, 5))
	studentNo := strings.TrimSpace(getCSVValue(record, 6))
	teacherNo := strings.TrimSpace(getCSVValue(record, 7))
	name := strings.TrimSpace(getCSVValue(record, 8))

	if username == "" {
		return false, fmt.Errorf("username 不能为空")
	}
	if role == "" {
		role = model.RoleStudent
	}
	if status == "" {
		status = model.UserStatusActive
	}

	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil && !errors.Is(err, identitycontracts.ErrUserNotFound) {
		return false, err
	}

	if existing == nil || errors.Is(err, identitycontracts.ErrUserNotFound) {
		if password == "" {
			return false, fmt.Errorf("新用户必须提供 password")
		}
		_, createErr := s.CreateUser(ctx, identitycontracts.CreateUserInput{
			Username:  username,
			Name:      name,
			Password:  password,
			Email:     email,
			StudentNo: studentNo,
			TeacherNo: teacherNo,
			ClassName: className,
			Role:      role,
			Status:    status,
		})
		if createErr != nil {
			return false, createErr
		}
		return true, nil
	}

	req := identitycontracts.UpdateUserInput{
		Name:      &name,
		Email:     &email,
		StudentNo: &studentNo,
		TeacherNo: &teacherNo,
		ClassName: &className,
		Role:      &role,
		Status:    &status,
	}
	if password != "" {
		req.Password = &password
	}

	_, updateErr := s.UpdateUser(ctx, existing.ID, req)
	return false, updateErr
}

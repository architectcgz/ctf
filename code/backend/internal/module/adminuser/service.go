package adminuser

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Service struct {
	repo       *Repository
	pagination config.PaginationConfig
	log        *zap.Logger
}

func NewService(repo *Repository, pagination config.PaginationConfig, log *zap.Logger) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &Service{
		repo:       repo,
		pagination: pagination,
		log:        log,
	}
}

func (s *Service) ListUsers(ctx context.Context, query *dto.AdminUserQuery) ([]dto.AdminUserResp, int64, int, int, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = s.pagination.DefaultPageSize
	}
	if size > s.pagination.MaxPageSize {
		size = s.pagination.MaxPageSize
	}

	users, total, err := s.repo.List(ctx, UserListFilter{
		Keyword:   strings.TrimSpace(query.Keyword),
		StudentNo: strings.TrimSpace(query.StudentNo),
		TeacherNo: strings.TrimSpace(query.TeacherNo),
		Role:      strings.TrimSpace(query.Role),
		Status:    strings.TrimSpace(query.Status),
		ClassName: strings.TrimSpace(query.ClassName),
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}

	items := make([]dto.AdminUserResp, 0, len(users))
	for _, user := range users {
		items = append(items, toAdminUserResp(user))
	}
	return items, total, page, size, nil
}

func (s *Service) CreateUser(ctx context.Context, req *dto.CreateAdminUserReq) (*dto.AdminUserResp, error) {
	identity := normalizeIdentityNumbers(req.Role, req.StudentNo, req.TeacherNo)
	user := &model.User{
		Username:  strings.TrimSpace(req.Username),
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

func (s *Service) UpdateUser(ctx context.Context, userID int64, req *dto.UpdateAdminUserReq) (*dto.AdminUserResp, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, mapServiceError(err)
	}

	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		if err := user.SetPassword(strings.TrimSpace(*req.Password)); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
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

func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.repo.Delete(ctx, userID); err != nil {
		return mapServiceError(err)
	}
	return nil
}

func (s *Service) ImportUsers(ctx context.Context, reader io.Reader) (*dto.ImportUsersResp, error) {
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

func (s *Service) importRow(ctx context.Context, record []string) (bool, error) {
	username := strings.TrimSpace(getCSVValue(record, 0))
	password := strings.TrimSpace(getCSVValue(record, 1))
	email := strings.TrimSpace(getCSVValue(record, 2))
	className := strings.TrimSpace(getCSVValue(record, 3))
	role := strings.TrimSpace(getCSVValue(record, 4))
	status := strings.TrimSpace(getCSVValue(record, 5))
	studentNo := strings.TrimSpace(getCSVValue(record, 6))
	teacherNo := strings.TrimSpace(getCSVValue(record, 7))

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
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return false, err
	}

	if existing == nil || errors.Is(err, ErrUserNotFound) {
		if password == "" {
			return false, fmt.Errorf("新用户必须提供 password")
		}
		_, createErr := s.CreateUser(ctx, &dto.CreateAdminUserReq{
			Username:  username,
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

	req := &dto.UpdateAdminUserReq{
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

func toAdminUserResp(user *model.User) dto.AdminUserResp {
	var email *string
	if strings.TrimSpace(user.Email) != "" {
		email = &user.Email
	}
	var studentNo *string
	if strings.TrimSpace(user.StudentNo) != "" {
		studentNo = &user.StudentNo
	}
	var teacherNo *string
	if strings.TrimSpace(user.TeacherNo) != "" {
		teacherNo = &user.TeacherNo
	}
	var className *string
	if strings.TrimSpace(user.ClassName) != "" {
		className = &user.ClassName
	}
	updatedAt := user.UpdatedAt
	return dto.AdminUserResp{
		ID:        user.ID,
		Username:  user.Username,
		Email:     email,
		StudentNo: studentNo,
		TeacherNo: teacherNo,
		ClassName: className,
		Status:    user.Status,
		Roles:     []string{user.Role},
		CreatedAt: user.CreatedAt,
		UpdatedAt: &updatedAt,
	}
}

func defaultUserStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return model.UserStatusActive
	}
	return strings.TrimSpace(status)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return errcode.ErrNotFound
	case errors.Is(err, ErrUsernameExists):
		return errcode.ErrUsernameExists
	case errors.Is(err, ErrEmailExists):
		return errcode.ErrEmailExists
	case errors.Is(err, ErrStudentNoExists):
		return errcode.ErrStudentNoExists
	case errors.Is(err, ErrTeacherNoExists):
		return errcode.ErrTeacherNoExists
	case errors.Is(err, ErrRoleNotFound):
		return errcode.ErrInternal.WithCause(err)
	default:
		return errcode.ErrInternal.WithCause(err)
	}
}

func looksLikeHeader(record []string) bool {
	return strings.EqualFold(strings.TrimSpace(getCSVValue(record, 0)), "username")
}

func isBlankRecord(record []string) bool {
	for _, item := range record {
		if strings.TrimSpace(item) != "" {
			return false
		}
	}
	return true
}

func getCSVValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return record[index]
}

type identityNumbers struct {
	StudentNo string
	TeacherNo string
}

func normalizeIdentityNumbers(role, studentNo, teacherNo string) identityNumbers {
	normalized := identityNumbers{
		StudentNo: strings.TrimSpace(studentNo),
		TeacherNo: strings.TrimSpace(teacherNo),
	}

	switch strings.TrimSpace(role) {
	case model.RoleStudent:
		normalized.TeacherNo = ""
	case model.RoleTeacher:
		normalized.StudentNo = ""
	default:
		normalized.StudentNo = ""
		normalized.TeacherNo = ""
	}

	return normalized
}

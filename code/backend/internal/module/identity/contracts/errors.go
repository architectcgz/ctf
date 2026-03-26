package contracts

import "errors"

var ErrUserNotFound = errors.New("identity user not found")
var ErrUsernameExists = errors.New("identity username already exists")
var ErrEmailExists = errors.New("identity email already exists")
var ErrStudentNoExists = errors.New("identity student no already exists")
var ErrTeacherNoExists = errors.New("identity teacher no already exists")
var ErrRoleNotFound = errors.New("identity role not found")

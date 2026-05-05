package commands

type RegisterInput struct {
	Username  string
	Password  string
	Email     string
	ClassName string
}

type LoginInput struct {
	Username string
	Password string
}

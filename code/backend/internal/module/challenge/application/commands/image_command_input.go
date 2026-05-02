package commands

type CreateImageInput struct {
	Name        string
	Tag         string
	Description string
}

type UpdateImageInput struct {
	Description *string
	Status      string
}

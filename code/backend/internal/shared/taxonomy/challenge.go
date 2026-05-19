package taxonomy

const (
	DimensionWeb       = "web"
	DimensionPwn       = "pwn"
	DimensionReverse   = "reverse"
	DimensionCrypto    = "crypto"
	DimensionMisc      = "misc"
	DimensionForensics = "forensics"
)

var ValidDimensions = map[string]bool{
	DimensionWeb:       true,
	DimensionPwn:       true,
	DimensionReverse:   true,
	DimensionCrypto:    true,
	DimensionMisc:      true,
	DimensionForensics: true,
}

var AllDimensions = []string{
	DimensionWeb,
	DimensionPwn,
	DimensionReverse,
	DimensionCrypto,
	DimensionMisc,
	DimensionForensics,
}

const (
	DifficultyBeginner = "beginner"
	DifficultyEasy     = "easy"
	DifficultyMedium   = "medium"
	DifficultyHard     = "hard"
	DifficultyInsane   = "insane"
)

func IsValidDimension(dimension string) bool {
	return ValidDimensions[dimension]
}

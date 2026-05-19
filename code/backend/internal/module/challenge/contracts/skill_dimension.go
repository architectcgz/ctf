package contracts

// 能力维度枚举
const (
	DimensionWeb       = "web"
	DimensionPwn       = "pwn"
	DimensionReverse   = "reverse"
	DimensionCrypto    = "crypto"
	DimensionMisc      = "misc"
	DimensionForensics = "forensics"
)

// ValidDimensions 合法维度集合
var ValidDimensions = map[string]bool{
	DimensionWeb:       true,
	DimensionPwn:       true,
	DimensionReverse:   true,
	DimensionCrypto:    true,
	DimensionMisc:      true,
	DimensionForensics: true,
}

// AllDimensions 所有维度列表
var AllDimensions = []string{
	DimensionWeb,
	DimensionPwn,
	DimensionReverse,
	DimensionCrypto,
	DimensionMisc,
	DimensionForensics,
}

// IsValidDimension 检查维度是否合法
func IsValidDimension(dimension string) bool {
	return ValidDimensions[dimension]
}

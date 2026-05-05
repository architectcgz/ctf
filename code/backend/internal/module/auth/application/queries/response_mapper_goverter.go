package queries

import "ctf-platform/internal/dto"

type casStatusSource struct {
	Provider      string
	Enabled       bool
	Configured    bool
	AutoProvision bool
	LoginPath     string
	CallbackPath  string
}

type casLoginSource struct {
	Provider    string
	RedirectURL string
	CallbackURL string
}

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type authQueryResponseMapper interface {
	ToCASStatusResp(source casStatusSource) dto.CASStatusResp
	ToCASStatusRespPtr(source casStatusSource) *dto.CASStatusResp

	ToCASLoginResp(source casLoginSource) dto.CASLoginResp
	ToCASLoginRespPtr(source casLoginSource) *dto.CASLoginResp
}

var authQueryResponseMapperInst authQueryResponseMapper

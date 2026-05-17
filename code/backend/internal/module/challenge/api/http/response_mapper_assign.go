//go:build !goverter

package http

func init() {
	challengeResponseMapper = &ChallengeResponseMapperImpl{}
}

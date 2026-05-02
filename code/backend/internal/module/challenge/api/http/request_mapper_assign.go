//go:build !goverter

package http

func init() {
	challengeRequestMapper = &ChallengeRequestMapperImpl{}
}

//go:build !goverter

package http

func init() {
	notificationResponseMapper = &notificationHTTPResponseMapperImpl{}
}

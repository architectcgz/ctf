//go:build !goverter

package commands

func init() {
	notificationMapper = &notificationResponseMapperImpl{}
}

//go:build !goverter

package queries

func init() {
	notificationMapper = &notificationResponseMapperImpl{}
}

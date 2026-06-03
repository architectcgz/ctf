package app

import (
	"net/http"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	systemapp "ctf-platform/internal/testutil/systemapp"
)

func newPracticeFlowTestConfig(t *testing.T) *config.Config {
	t.Helper()
	return systemapp.NewPracticeFlowTestConfig(t)
}

func createFlowImage(t *testing.T, db *gorm.DB) *appImageRow {
	t.Helper()

	image := &appImageRow{
		Name:   "ctf/web-basic",
		Tag:    "v1",
		Status: challengecontracts.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image
}

func loginForSession(t *testing.T, router http.Handler, username, password string) *http.Cookie {
	t.Helper()
	return systemapp.LoginForSession(t, router, username, password)
}

func sessionHeaders(cookie *http.Cookie) map[string]string {
	return systemapp.SessionHeaders(cookie)
}

func loginForToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	return systemapp.LoginForToken(t, router, username, password)
}

func bearerHeaders(token string) map[string]string {
	return systemapp.BearerHeaders(token)
}

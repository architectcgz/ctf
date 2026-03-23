package composition

import (
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	"ctf-platform/internal/module/identity"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identityapp "ctf-platform/internal/module/identity/application"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	jwtpkg "ctf-platform/pkg/jwt"
)

type IdentityModule struct {
	AdminHandler   *identityhttp.Handler
	ProfileService identity.ProfileService
	TokenService   identity.Authenticator

	users identity.UserRepository
}

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	jwtManager, err := jwtpkg.NewManager(cfg.Auth, cfg.App.Name)
	if err != nil {
		return nil, err
	}

	users := identityinfra.NewRepository(db)
	tokenService := identity.NewModule(authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, cache, jwtManager))
	adminService := identityapp.NewAdminService(users, cfg.Pagination, log.Named("identity_admin_service"))
	profileService := identityapp.NewProfileService(users, log.Named("identity_profile_service"))

	return &IdentityModule{
		AdminHandler:   identityhttp.NewHandler(adminService),
		ProfileService: profileService,
		TokenService:   tokenService,
		users:          users,
	}, nil
}

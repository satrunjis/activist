package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	appauth "activist-base/src/application/auth"
	appauthorization "activist-base/src/application/authorization"
	appeventlog "activist-base/src/application/eventlog"
	appsearch "activist-base/src/application/search"
	appuser "activist-base/src/application/user"
	"activist-base/src/config"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres"
	httpauth "activist-base/src/transport/http/auth"
	httpdivision "activist-base/src/transport/http/division"
	httpeventlog "activist-base/src/transport/http/eventlog"
	httpmembership "activist-base/src/transport/http/membership"
	httpposition "activist-base/src/transport/http/position"
	httprole "activist-base/src/transport/http/role"
	httpsearch "activist-base/src/transport/http/search"
	httpsession "activist-base/src/transport/http/session"
	httpuser "activist-base/src/transport/http/user"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	Config config.Config
	Logger *slog.Logger
	Store  *postgres.Store

	AuthService          httpauth.Service
	AuthorizationService AuthorizationService
	SessionRepository    httpsession.Repository
	UserProfileStore     httpuser.ProfileRepository
	DivisionStore        httpdivision.Repository
	RoleStore            httprole.Repository
	PositionStore        httpposition.Repository
	MembershipStore      httpmembership.Repository
	MembershipReader     httpuser.MembershipReader
	EventLogStore        httpeventlog.Repository
}

type AuthorizationService interface {
	ResolveForDivision(ctx context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error)
	ResolveForUser(ctx context.Context, actorID shared.UserID, targetUserID shared.UserID) (domainmembership.EffectivePermissions, error)
	ResolveGlobal(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error)
}

func NewRouter(deps Deps) http.Handler {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(newCORSMiddleware(corsOptions{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:4173",
			"http://127.0.0.1:4173",
		},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	cookieName := deps.Config.SessionCookieName
	if cookieName == "" {
		cookieName = "__Host-session"
	}

	authService := deps.AuthService
	authorizationService := deps.AuthorizationService
	sessionRepository := deps.SessionRepository
	userProfileStore := deps.UserProfileStore
	divisionStore := deps.DivisionStore
	roleStore := deps.RoleStore
	positionStore := deps.PositionStore
	membershipStore := deps.MembershipStore
	membershipReader := deps.MembershipReader
	eventLogStore := deps.EventLogStore
	var searchService *appsearch.Service
	var commandService *appeventlog.CommandService

	if deps.Store != nil && deps.Store.Queries != nil {
		authRepository := postgres.NewAuthRepository(deps.Store.Queries)
		eventLogRepo := postgres.NewEventLogRepository(deps.Store.Queries)
		commandService = appeventlog.NewCommandService(postgres.NewTxManager(deps.Store.Pool), eventLogRepo)
		if authService == nil {
			authService = appauth.NewService(appauth.Config{
				SessionIdleTTL:     deps.Config.SessionIdleTTL,
				SessionAbsoluteTTL: deps.Config.SessionAbsoluteTTL,
			}, authRepository)
		}
		if sessionRepository == nil {
			sessionRepository = authRepository
		}
		if authorizationService == nil {
			authorizationService = appauthorization.NewService(postgres.NewAuthorizationRepository(deps.Store.Queries))
		}
		if userProfileStore == nil {
			userProfileStore = postgres.NewUserProfileRepository(deps.Store.Queries)
		}
		if divisionStore == nil {
			divisionStore = postgres.NewDivisionRepository(deps.Store.Queries)
		}
		if roleStore == nil {
			roleStore = postgres.NewRoleRepository(deps.Store.Queries)
		}
		if positionStore == nil {
			positionStore = postgres.NewPositionRepository(deps.Store.Queries, deps.Store.Pool)
		}
		membershipRepo := postgres.NewMembershipRepository(deps.Store.Queries, deps.Store.Pool)
		if membershipStore == nil {
			membershipStore = membershipRepo
		}
		if membershipReader == nil {
			membershipReader = membershipRepo
		}
		if eventLogStore == nil {
			eventLogStore = eventLogRepo
		}
		if authorizationService != nil {
			searchService = appsearch.NewService(
				searchRepositoryAdapter{repo: postgres.NewSearchRepository(deps.Store.Queries)},
				authorizationService,
			)
		}
	}
	if membershipReader == nil {
		membershipReader = emptyMembershipReader{}
	}

	r.Route("/api/v1", func(api chi.Router) {
		authHandler := httpauth.NewHandler(authService, cookieName)
		authHandler.SetAuthorizationService(authorizationService)
		authMiddleware := httpsession.NewMiddleware(sessionRepository, httpsession.Options{
			CookieName:     cookieName,
			SessionIdleTTL: deps.Config.SessionIdleTTL,
			DevAssumeAdmin: deps.Config.DevAuthAssumeAdmin,
			DevAdminLogin:  deps.Config.DevAuthAdminLogin,
		})
		httpauth.RegisterRoutes(api, authHandler, authMiddleware)
		if userProfileStore != nil {
			userHandler := httpuser.NewHandler(userProfileStore, membershipReader, authorizationService)
			httpuser.RegisterRoutes(api, userHandler, authMiddleware)
		}
		if divisionStore != nil {
			divisionHandler := httpdivision.NewHandler(divisionStore, authorizationService, commandService)
			httpdivision.RegisterRoutes(api, divisionHandler, authMiddleware)
		}
		if roleStore != nil {
			roleHandler := httprole.NewHandler(roleStore, authorizationService, commandService)
			httprole.RegisterRoutes(api, roleHandler, authMiddleware)
		}
		if positionStore != nil {
			positionHandler := httpposition.NewHandler(positionStore, authorizationService, commandService)
			httpposition.RegisterRoutes(api, positionHandler, authMiddleware)
		}
		if membershipStore != nil {
			membershipHandler := httpmembership.NewHandler(membershipStore, authorizationService, commandService)
			httpmembership.RegisterRoutes(api, membershipHandler, authMiddleware)
		}
		if eventLogStore != nil {
			eventLogHandler := httpeventlog.NewHandler(eventLogStore, authorizationService)
			httpeventlog.RegisterRoutes(api, eventLogHandler, authMiddleware)
		}
		if searchService != nil {
			searchHandler := httpsearch.NewHandler(searchService)
			httpsearch.RegisterRoutes(api, searchHandler, authMiddleware)
		}
	})

	return r
}

type emptyMembershipReader struct{}

func (emptyMembershipReader) ListByUser(_ context.Context, _ shared.UserID) ([]appuser.MembershipView, error) {
	return []appuser.MembershipView{}, nil
}

func stringPtr(value string) *string {
	return &value
}

type searchRepositoryAdapter struct {
	repo *postgres.SearchRepository
}

func (a searchRepositoryAdapter) SearchUsers(
	ctx context.Context,
	params appsearch.SearchUsersParams,
) (appsearch.SearchUsersResult, error) {
	repoResult, err := a.repo.SearchUsers(ctx, postgres.SearchUsersParams{
		FirstName:       params.FirstName,
		LastName:        params.LastName,
		MiddleName:      params.MiddleName,
		Login:           params.Login,
		GroupNumber:     params.GroupNumber,
		Institute:       params.Institute,
		About:           params.About,
		PositionTitle:   params.PositionTitle,
		RoleName:        params.RoleName,
		IncludeArchived: params.IncludeArchived,
		Limit:           params.Limit,
		Offset:          params.Offset,
	})
	if err != nil {
		return appsearch.SearchUsersResult{}, err
	}

	items := make([]appsearch.SearchUser, 0, len(repoResult.Items))
	for _, item := range repoResult.Items {
		var links []shared.Link
		if len(item.SocialLinks) > 0 {
			if err := json.Unmarshal(item.SocialLinks, &links); err != nil {
				return appsearch.SearchUsersResult{}, err
			}
		}
		var birthDate *string
		if item.BirthDate != nil {
			value := item.BirthDate.UTC().Format("2006-01-02")
			birthDate = &value
		}
		var socialLinks *[]shared.Link
		if links != nil {
			cloned := append([]shared.Link(nil), links...)
			socialLinks = &cloned
		}

		items = append(items, appsearch.SearchUser{
			ID:              shared.UserID(item.ID),
			FirstName:       item.FirstName,
			LastName:        item.LastName,
			MiddleName:      item.MiddleName,
			Login:           stringPtr(item.Login),
			GroupNumber:     item.GroupNumber,
			Institute:       item.Institute,
			About:           item.About,
			Phone:           item.Phone,
			SocialLinks:     socialLinks,
			BirthDate:       birthDate,
			GradebookNumber: item.GradebookNumber,
		})
	}

	return appsearch.SearchUsersResult{
		Items: items,
		Total: repoResult.Total,
	}, nil
}

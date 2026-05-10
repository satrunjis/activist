package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "activist-base/src/application/auth"
	domainauth "activist-base/src/domain/auth"
	"activist-base/src/domain/shared"
	pgrepo "activist-base/src/repository/postgres"
	testpostgres "activist-base/src/testutil/postgres"
)

func TestRegister(t *testing.T) {

	ctx := context.Background()
	tdb, cleanupDB, err := testpostgres.NewTestDatabase(ctx)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	t.Cleanup(cleanupDB)

	store, err := pgrepo.New(ctx, tdb.DSN)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	repo := pgrepo.NewAuthRepository(store.Queries)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	svc := appauth.NewService(appauth.Config{SessionIdleTTL: 2 * time.Hour, SessionAbsoluteTTL: 24 * time.Hour}, repo)
	svc.SetNowFunc(func() time.Time { return now })

	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)
	registerResult, err := svc.Register(ctx, appauth.RegisterInput{
		Login:           "  activist_user  ",
		Password:        "StrongPassw0rd!",
		FirstName:       "Anna",
		GradebookNumber: "GB-001",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if registerResult.User.Login != "activist_user" {
		t.Fatalf("expected trimmed login, got %q", registerResult.User.Login)
	}
	if registerResult.Session.Token == "" {
		t.Fatal("expected opaque session token")
	}

	persistedUser, err := store.Queries.GetUserByLogin(ctx, "activist_user")
	if err != nil {
		t.Fatalf("get user by login: %v", err)
	}
	if persistedUser.PasswordHash == "" {
		t.Fatal("expected password hash to be stored")
	}
	if persistedUser.PasswordHash == "StrongPassw0rd!" {
		t.Fatal("raw password was stored")
	}

	sessionByHash, err := store.Queries.GetSessionByTokenHash(ctx, domainauth.HashToken(registerResult.Session.Token))
	if err != nil {
		t.Fatalf("get session by token hash: %v", err)
	}
	if sessionByHash.TokenHash == registerResult.Session.Token {
		t.Fatal("raw session token was stored")
	}
}

func TestRegisterRejectsDuplicateLoginAndGradebook(t *testing.T) {

	ctx := context.Background()
	tdb, cleanupDB, err := testpostgres.NewTestDatabase(ctx)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	t.Cleanup(cleanupDB)

	store, err := pgrepo.New(ctx, tdb.DSN)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	repo := pgrepo.NewAuthRepository(store.Queries)
	svc := appauth.NewService(appauth.Config{SessionIdleTTL: 2 * time.Hour, SessionAbsoluteTTL: 24 * time.Hour}, repo)
	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)

	_, err = svc.Register(ctx, appauth.RegisterInput{
		Login:           "duplicate_user",
		Password:        "StrongPassw0rd!",
		FirstName:       "Anna",
		GradebookNumber: "GB-777",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	if err != nil {
		t.Fatalf("seed register: %v", err)
	}

	_, err = svc.Register(ctx, appauth.RegisterInput{
		Login:           "duplicate_user",
		Password:        "StrongPassw0rd!",
		FirstName:       "Bob",
		GradebookNumber: "GB-778",
		GroupNumber:     "B-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	var loginTaken *shared.Error
	if !errors.As(err, &loginTaken) || loginTaken.Code != "validation.login_taken" {
		t.Fatalf("expected validation.login_taken, got %v", err)
	}

	_, err = svc.Register(ctx, appauth.RegisterInput{
		Login:           "other_user",
		Password:        "StrongPassw0rd!",
		FirstName:       "Bob",
		GradebookNumber: "GB-777",
		GroupNumber:     "B-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	var gradebookTaken *shared.Error
	if !errors.As(err, &gradebookTaken) || gradebookTaken.Code != "validation.gradebook_number_taken" {
		t.Fatalf("expected validation.gradebook_number_taken, got %v", err)
	}
}

func TestLoginInvalidCredentialsParity(t *testing.T) {

	ctx := context.Background()
	tdb, cleanupDB, err := testpostgres.NewTestDatabase(ctx)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	t.Cleanup(cleanupDB)

	store, err := pgrepo.New(ctx, tdb.DSN)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	repo := pgrepo.NewAuthRepository(store.Queries)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	svc := appauth.NewService(appauth.Config{SessionIdleTTL: 2 * time.Hour, SessionAbsoluteTTL: 24 * time.Hour}, repo)
	svc.SetNowFunc(func() time.Time { return now })

	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)
	_, err = svc.Register(ctx, appauth.RegisterInput{
		Login:           "known_user",
		Password:        "StrongPassw0rd!",
		FirstName:       "Anna",
		GradebookNumber: "GB-001",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, unknownErr := svc.Login(ctx, appauth.LoginInput{
		Login:     "missing_user",
		Password:  "StrongPassw0rd!",
		IP:        "127.0.0.1",
		UserAgent: "go-test",
	})
	if !errors.Is(unknownErr, appauth.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for unknown login, got %v", unknownErr)
	}

	_, wrongPasswordErr := svc.Login(ctx, appauth.LoginInput{
		Login:     "known_user",
		Password:  "wrong",
		IP:        "127.0.0.1",
		UserAgent: "go-test",
	})
	if !errors.Is(wrongPasswordErr, appauth.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for wrong password, got %v", wrongPasswordErr)
	}

	if unknownErr != wrongPasswordErr {
		t.Fatalf("expected same invalid-credentials error path, got %v and %v", unknownErr, wrongPasswordErr)
	}
}

func TestSessionLifecycle(t *testing.T) {

	ctx := context.Background()
	tdb, cleanupDB, err := testpostgres.NewTestDatabase(ctx)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	t.Cleanup(cleanupDB)

	store, err := pgrepo.New(ctx, tdb.DSN)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	repo := pgrepo.NewAuthRepository(store.Queries)
	base := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	now := base
	svc := appauth.NewService(appauth.Config{SessionIdleTTL: 20 * time.Minute, SessionAbsoluteTTL: 2 * time.Hour}, repo)
	svc.SetNowFunc(func() time.Time { return now })

	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)
	registered, err := svc.Register(ctx, appauth.RegisterInput{
		Login:           "session_user",
		Password:        "StrongPassw0rd!",
		FirstName:       "Anna",
		GradebookNumber: "GB-001",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	now = now.Add(10 * time.Minute)
	sessionView, err := svc.GetSession(ctx, appauth.GetSessionInput{Token: registered.Session.Token})
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if !sessionView.Session.IdleExpiresAt.After(now) {
		t.Fatalf("expected idle expiry to move forward, got %v", sessionView.Session.IdleExpiresAt)
	}

	now = base.Add(31 * time.Minute)
	_, err = svc.GetSession(ctx, appauth.GetSessionInput{Token: registered.Session.Token})
	if !errors.Is(err, appauth.ErrInvalidSession) {
		t.Fatalf("expected invalid session after idle ttl boundary, got %v", err)
	}

	loggedIn, err := svc.Login(ctx, appauth.LoginInput{
		Login:     "session_user",
		Password:  "StrongPassw0rd!",
		IP:        "127.0.0.1",
		UserAgent: "go-test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	now = base.Add(2 * time.Hour)
	_, err = svc.GetSession(ctx, appauth.GetSessionInput{Token: loggedIn.Session.Token})
	if !errors.Is(err, appauth.ErrInvalidSession) {
		t.Fatalf("expected invalid session at absolute ttl boundary, got %v", err)
	}

	now = base.Add(3 * time.Hour)
	_, err = svc.GetSession(ctx, appauth.GetSessionInput{Token: registered.Session.Token})
	if !errors.Is(err, appauth.ErrInvalidSession) {
		t.Fatalf("expected invalid session after absolute ttl, got %v", err)
	}
}

func TestLogout(t *testing.T) {

	ctx := context.Background()
	tdb, cleanupDB, err := testpostgres.NewTestDatabase(ctx)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	t.Cleanup(cleanupDB)

	store, err := pgrepo.New(ctx, tdb.DSN)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	repo := pgrepo.NewAuthRepository(store.Queries)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	svc := appauth.NewService(appauth.Config{SessionIdleTTL: 2 * time.Hour, SessionAbsoluteTTL: 24 * time.Hour}, repo)
	svc.SetNowFunc(func() time.Time { return now })

	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)
	registered, err := svc.Register(ctx, appauth.RegisterInput{
		Login:           "logout_user",
		Password:        "StrongPassw0rd!",
		FirstName:       "Anna",
		GradebookNumber: "GB-001",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       birthDate,
		IP:              "127.0.0.1",
		UserAgent:       "go-test",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	loggedIn, err := svc.Login(ctx, appauth.LoginInput{
		Login:     "logout_user",
		Password:  "StrongPassw0rd!",
		IP:        "127.0.0.1",
		UserAgent: "go-test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if err := svc.Logout(ctx, appauth.LogoutInput{Token: registered.Session.Token}); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = svc.GetSession(ctx, appauth.GetSessionInput{Token: registered.Session.Token})
	if !errors.Is(err, appauth.ErrInvalidSession) {
		t.Fatalf("expected invalid session after logout, got %v", err)
	}

	if err := svc.LogoutAll(ctx, appauth.LogoutAllInput{UserID: loggedIn.User.ID}); err != nil {
		t.Fatalf("logout all: %v", err)
	}

	_, err = svc.GetSession(ctx, appauth.GetSessionInput{Token: loggedIn.Session.Token})
	if !errors.Is(err, appauth.ErrInvalidSession) {
		t.Fatalf("expected invalid session after logout all, got %v", err)
	}
}

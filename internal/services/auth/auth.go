package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	jwt "sso/internal/lib/loger/JWT"
	sl "sso/internal/lib/loger/slog"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash string,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, uid int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appId int64) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {

	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int64,
) (string, error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.Int64("appId", appId),
	)
	log.Info("Login", "email", email, "appId", appId)

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", "error", err)
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password)); err != nil {
		a.log.Info("invalid password", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	app, err := a.appProvider.App(ctx, appId)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to create token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil

}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("RegisterNewUser", "email", email)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := a.userSaver.SaveUser(ctx, email, string(passHash))
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", "error", err)
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registered")
	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userId int64,
) (bool, error) {
	const op = "auth.IsAdmin"
	log := a.log.With(
		slog.String("op", op),
		slog.Int64("userId", userId),
	)
	log.Info("IsAdmin", "userId", userId)

	isAdmin, err := a.userProvider.IsAdmin(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", "error", err)
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("checked if user is admin", slog.Bool("isAdmin", isAdmin))
	return isAdmin, nil
}

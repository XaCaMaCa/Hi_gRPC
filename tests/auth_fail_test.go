package tests

import (
	suite "sso/tests/migration/SUITE.GO"
	"testing"

	authv1 "github.com/XaCaMaCa/protos/gen/go"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestRegisterFailCases проверяет ошибки при регистрации
func TestRegisterFailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr codes.Code
	}{
		{
			name:        "Регистрация с пустым email",
			email:       "",
			password:    generatePassword(),
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Регистрация с пустым паролем",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Регистрация с пустыми данными",
			email:       "",
			password:    "",
			expectedErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.expectedErr, st.Code())
		})
	}
}

// TestRegisterDuplicateEmail проверяет повторную регистрацию с тем же email
func TestRegisterDuplicateEmail(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := generatePassword()

	// Первая регистрация должна пройти успешно
	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	// Повторная регистрация с тем же email должна вернуть ошибку
	_, err = st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)

	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, grpcStatus.Code())
}

// TestLoginFailCases проверяет ошибки при входе
func TestLoginFailCases(t *testing.T) {
	ctx, st := suite.New(t)

	// Сначала регистрируем пользователя для некоторых тестов
	email := gofakeit.Email()
	password := generatePassword()
	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectedErr codes.Code
	}{
		{
			name:        "Вход с пустым email",
			email:       "",
			password:    password,
			appId:       appId,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Вход с пустым паролем",
			email:       email,
			password:    "",
			appId:       appId,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Вход с пустым app_id",
			email:       email,
			password:    password,
			appId:       emptyAppId,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Вход с неправильным паролем",
			email:       email,
			password:    "wrong_password_123",
			appId:       appId,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Вход несуществующего пользователя",
			email:       "nonexistent@example.com",
			password:    password,
			appId:       appId,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Вход с несуществующим app_id",
			email:       email,
			password:    password,
			appId:       999,
			expectedErr: codes.Internal, // или NotFound в зависимости от реализации
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appId,
			})
			require.Error(t, err)

			grpcStatus, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.expectedErr, grpcStatus.Code())
		})
	}
}

// TestIsAdminFailCases проверяет ошибки при проверке админа
func TestIsAdminFailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		userId      int64
		expectedErr codes.Code
	}{
		{
			name:        "Проверка с пустым user_id",
			userId:      0,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "Проверка несуществующего пользователя",
			userId:      999999,
			expectedErr: codes.Internal, // Сервер возвращает Internal при отсутствии пользователя
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.IsAdmin(ctx, &authv1.IsAdminRequest{
				UserId: tt.userId,
			})
			require.Error(t, err)

			grpcStatus, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.expectedErr, grpcStatus.Code())
		})
	}
}

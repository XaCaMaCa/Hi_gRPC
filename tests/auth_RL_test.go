package tests

import (
	suite "sso/tests/migration/SUITE.GO"
	"testing"
	"time"

	authv1 "github.com/XaCaMaCa/protos/gen/go"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppId  = 0
	appId       = 1
	appSecret   = "test-secret"
	passDefault = 10
)

func TestAuthRL(t *testing.T) {
	ctx, suite := suite.New(t)

	email := gofakeit.Email()
	password := generatePassword()
	respReg, err := suite.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := suite.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appId,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["user_id"].(float64)))
	assert.Equal(t, email, claims["email"])
	assert.Equal(t, appId, int(claims["app_id"].(float64)))
	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(suite.Config.TokenTTL).Unix(), int64(claims["exp"].(float64)), deltaSeconds)

}

func generatePassword() string {
	return gofakeit.Password(true, true, true, true, true, passDefault)
}

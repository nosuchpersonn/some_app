package app_test

import (
	"app"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Find(sql string) *app.User {
	return m.Called(sql).Get(0).(*app.User)
}

func (m *UserRepositoryMock) Save(user *app.User) {
	m.Called(user)
}

func TestUserApi_ProcessRequest(t *testing.T) {
	g := gin.Default()
	g.Use(func(c *gin.Context) {
		c.Set("auth_user_id", 999)
	})

	user := &app.User{
		Id:           999,
		Name:         "Alex",
		Phone:        "79222222222",
		IsAdmin:      false,
		LastViewedAt: time.Now().Add(-time.Hour * 48),
	}

	repo := &UserRepositoryMock{}
	repo.On("Find", mock.Anything).Return(user).Twice()
	repo.On("Save", user).Once()

	api := app.NewUserApi(repo)
	g.POST("/user", api.ProcessRequest)

	req, _ := http.NewRequest("POST", "/user?id=999", strings.NewReader(`{"name": "Petr", "phone": "79111111111"}`))
	res := httptest.NewRecorder()
	g.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var body map[string]string
	json.Unmarshal(res.Body.Bytes(), &body)
	require.Equal(t, "Petr", body["name"])
	require.Equal(t, "79111111111", body["phone"])

	require.Equal(t, "Petr", user.Name)
	require.Equal(t, "79111111111", user.Phone)
	require.WithinDuration(t, time.Now(), user.LastViewedAt, time.Second)
}

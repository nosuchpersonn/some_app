package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func validateName(name string) error {
	if len(name) > 10 {
		return errors.New("invalid name length")
	}
	return nil
}

func validatePhone(phone string) error {
	if len(phone) != 11 {
		return errors.New("invalid phone format")
	}
	return nil
}

type User struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	IsAdmin      bool      `json:"isAdmin"`
	LastViewedAt time.Time `json:"lastViewedAt"`
}

func (u *User) SetLastViewedAt(t time.Time) {
	u.LastViewedAt = t
}

type UserRepository interface {
	Find(sql string) *User
	Save(user *User)
}

type UserApi struct {
	userRepo UserRepository
	logger   *zap.Logger
}

func NewUserApi(userRepo UserRepository) *UserApi {
	api := &UserApi{
		userRepo: userRepo,
	}
	if os.Getenv("LOGGER_PROD") == "1" {
		api.logger, _ = zap.NewProduction()
	} else {
		api.logger, _ = zap.NewDevelopment()
	}
	return api
}

func (api *UserApi) ProcessRequest(c *gin.Context) {
	authUserId := c.GetInt("auth_user_id")
	authUser := api.userRepo.Find(fmt.Sprintf("SELECT * FROM user WHERE id = %d", authUserId))

	userId := c.Query("id")
	user := api.userRepo.Find(fmt.Sprintf("SELECT * FROM user WHERE id = %s", userId))

	if authUser.IsAdmin || authUser.Id == user.Id {
		if user != nil && !user.IsAdmin {
			user.SetLastViewedAt(time.Now())
		}

		if c.Request.Method == "POST" {
			bytes, _ := io.ReadAll(c.Request.Body)
			var body map[string]string
			json.Unmarshal(bytes, &body)

			var errs map[string]string
			if err := validateName(body["name"]); err != nil {
				errs["name"] = err.Error()
			} else {
				user.Name = body["name"]
			}
			if err := validatePhone(body["phone"]); err != nil {
				errs["phone"] = err.Error()
			} else {
				user.Phone = body["phone"]
			}
			if len(errs) > 0 {
				c.JSON(http.StatusUnprocessableEntity, errs)
			}
		}

		api.userRepo.Save(user)
		api.logger.Debug("user saved", zap.Int("user_id", user.Id))

		c.JSON(http.StatusOK, user)
		return
	}

	c.Status(http.StatusForbidden)
}

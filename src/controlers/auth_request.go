package controlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	apiservices "github.com/shabs76/roro-local-server/api_services"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	usersdataservices "github.com/shabs76/roro-local-server/database/users_data_services"
	"github.com/shabs76/roro-local-server/specials"
)

func LoginUser(c *gin.Context) {
	var req users.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Invalid request body: " + err.Error(),
			Message: "Invalid data format was submited to the server",
			Adv:     "none",
		})
		return
	}

	// fect user data
	st, usersList := usersdataservices.SelectUserDetailsWithRolesPass("email = ? ", []any{req.Email})
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch user data: " + st.Data,
			Message: "An error occurred while fetching user data",
			Adv:     "none",
		})
		return
	}

	// The remote server decides. A user missing here (added after the last pull, or
	// left out of the remote users list) or whose password changed on the remote server
	// is stored once the remote server accepts the login.
	knownHere := len(usersList) > 0 && specials.ComparePasswordHash(req.Password, usersList[0].Password)

	// send login request to remote server
	logRespo, err := apiservices.RemoteLogin(req)
	if err != nil {
		slog.Error(err.Error())
		if errors.Is(err, apiservices.ErrRemoteLoginRefused) {
			c.JSON(http.StatusUnauthorized, constants.NormalResponse{
				State:   constants.ErrorState,
				Data:    "Invalid email or password",
				Message: "Authentication failed",
				Adv:     "none",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    fmt.Sprintf("failed to login to the remote server, due to %s", err.Error()),
			Message: "An error occurred while logging in to the remote server",
			Adv:     "none",
		})
		return
	}

	if !knownHere {
		if err := apiservices.SaveSignedInUser(logRespo, req.Password); err != nil {
			slog.Error("Login accepted by the remote server, but the user could not be stored", "email", req.Email, "error", err)
			c.JSON(http.StatusInternalServerError, constants.NormalResponse{
				State:   constants.ErrorState,
				Data:    "Failed to save user: " + err.Error(),
				Message: "An error occurred while saving the user",
				Adv:     "none",
			})
			return
		}
	}

	// insert login session to local db
	stSave := usersdataservices.UserlonginsData([]users.UserLoginResponse{logRespo})
	if stSave.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to save login session: " + stSave.Data,
			Message: "An error occurred while saving login session",
			Adv:     "none",
		})
		return
	}

	c.JSON(
		http.StatusOK,
		constants.NormalResponse{
			State: constants.SuccessState,
			Data: gin.H{
				"user":     logRespo.User,
				"loginId":  logRespo.LoginID,
				"loginKey": logRespo.LoginKey,
			},
			Message: "Login successful",
			Adv:     "none",
		},
	)

}

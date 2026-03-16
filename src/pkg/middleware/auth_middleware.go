package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
	usersdataservices "github.com/shabs76/roro-local-server/database/users_data_services"
)

// AuthSession holds the authentication headers and user ID to be used by controllers
type AuthSession struct {
	LogID  string
	LogKey string
	UserID string
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow specific paths to bypass authentication
		if strings.HasPrefix(c.Request.URL.Path, "/manifest/media/get/") {
			c.Next()
			return
		}

		logId := c.GetHeader("Log-Id")
		logKey := c.GetHeader("Log-Key")

		if logId == "" || logKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"state": constants.ErrorState,
				"data":  "Unauthorized: Missing login headers",
			})
			c.Abort()
			return
		}

		// Validate credentials against the database
		// We check for matching login_id, login_key and ensure the status is 'active'
		st, logins := usersdataservices.SelectUserLogins(
			"login_id = ? AND login_key = ? AND status = ?",
			[]any{logId, logKey, constants.StatusTypes.Active},
		)

		if st.State != constants.SuccessState {
			c.JSON(http.StatusInternalServerError, gin.H{
				"state": constants.ErrorState,
				"data":  "Internal Server Error: Failed to validate session",
			})
			c.Abort()
			return
		}

		if len(logins) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"state": constants.ErrorState,
				"data":  "Unauthorized: Invalid or expired session",
			})
			c.Abort()
			return
		}

		// now fetch user details base on login details
		stu, users := usersdataservices.SelectUserDetailsWithRolesPass(" `user_id` = ? AND `status` = ? ", []any{logins[0].UserID, constants.StatusTypes.Active})

		if stu.State != constants.SuccessState {
			c.AbortWithStatusJSON(http.StatusInternalServerError, stu)
			return
		} else if len(users) <= 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"state": constants.ErrorState, "data": "Your account is no longer active"})
			return
		}

		// Session is valid
		// Store UserID in context for use in controllers
		c.Set("userID", logins[0].UserID)

		// Forward headers/session as a single struct
		c.Set("authSession", AuthSession{
			LogID:  logId,
			LogKey: logKey,
			UserID: logins[0].UserID,
		})

		// send user details to the next function
		c.Set("user", users[0])

		c.Next()
	}
}

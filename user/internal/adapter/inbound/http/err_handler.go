package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/htan06/Moments/user/internal/domain"

	"github.com/gin-gonic/gin"
)

var businessError = map[error]int{
	domain.ErrUserNotFound:    http.StatusNotFound,
	domain.ErrUserIDInvalid:   http.StatusUnprocessableEntity,
	domain.ErrUsernameInvalid: http.StatusUnprocessableEntity,
	domain.ErrNameInvalid:     http.StatusUnprocessableEntity,
	domain.ErrNegativeCount:   http.StatusUnprocessableEntity,
}

func toHTTPCode(e error) int {
	for be, code := range businessError {
		if errors.Is(e, be) {
			return code
		}
	}
	return http.StatusInternalServerError
}

func HandleError(c *gin.Context, err error) {
	httpCode := toHTTPCode(err)
	if httpCode < 500 || httpCode >= 600 {
		c.JSON(httpCode, err.Error())
		return
	}

	fmt.Println("ERROR: ", err.Error())
	c.AbortWithStatus(http.StatusInternalServerError)
}

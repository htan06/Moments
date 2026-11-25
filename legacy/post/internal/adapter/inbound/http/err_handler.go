package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/post/internal/domain"
)

var businessError = map[error]int{
	domain.ErrUserNotFound:       http.StatusNotFound,
	domain.ErrMediaCountInvalid:  http.StatusUnprocessableEntity,
	domain.ErrAspectRatioInvalid: http.StatusUnprocessableEntity,
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

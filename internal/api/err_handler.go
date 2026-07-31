package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/errs"
)

var errorTypeTable = map[errs.ErrorType]int{
	errs.Invalid:               http.StatusUnprocessableEntity,
	errs.Incorrect:             http.StatusUnprocessableEntity,
	errs.Conflict:              http.StatusConflict,
	errs.NotFound:              http.StatusNotFound,
	errs.AuthenticationFailure: http.StatusUnauthorized,
}

func HandleError(c *gin.Context, err error) {
	if e, ok := errors.AsType[*errs.Error](err); ok {
		if httpCode, ok := errorTypeTable[e.Type]; ok {
			c.JSON(httpCode, gin.H{
				"error_code": e.Code,
			})
			return
		}
	}

	fmt.Println("ERROR: ", err.Error())
	c.AbortWithStatus(http.StatusInternalServerError)
}

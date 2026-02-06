package apperror

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
)

// HandleError converts an error to an appropriate JSON response.
// If the error is an *AppError, its Code/Err/Message are used.
// Otherwise a generic 500 response is returned.
func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, dto.ErrorResponse{
			Error:   appErr.Err,
			Message: appErr.Message,
			Code:    appErr.Code,
		})
		return
	}
	c.JSON(500, dto.ErrorResponse{
		Error:   "Internal Server Error",
		Message: err.Error(),
		Code:    500,
	})
}

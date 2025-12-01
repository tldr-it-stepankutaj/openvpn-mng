package logger

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger returns a gin.HandlerFunc (middleware) that logs requests using our logger
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		// Log based on status code
		if statusCode >= 500 {
			Error("HTTP request",
				"status", statusCode,
				"latency", latency,
				"client_ip", clientIP,
				"method", method,
				"path", path,
				"error", c.Errors.ByType(gin.ErrorTypePrivate).String(),
			)
		} else if statusCode >= 400 {
			Warn("HTTP request",
				"status", statusCode,
				"latency", latency,
				"client_ip", clientIP,
				"method", method,
				"path", path,
			)
		} else {
			Info("HTTP request",
				"status", statusCode,
				"latency", latency,
				"client_ip", clientIP,
				"method", method,
				"path", path,
			)
		}
	}
}

// GinRecovery returns a gin.HandlerFunc (middleware) that recovers from panics and logs them
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Error("Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

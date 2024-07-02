package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
	"simpleServer/pkg/logging"
	"simpleServer/pkg/trace"
	"time"
)

const (
	XRequestIdKey = "X-Request-ID"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		requestId := context.Request.Header.Get(XRequestIdKey)
		if requestId == "" {
			id, _ := uuid.NewV4()
			requestId = id.String()
		}

		context.Request = context.Request.WithContext(trace.WithRequestID(context, requestId))
		logger := logging.DefaultLogger().With("requestId", requestId)
		context.Request = context.Request.WithContext(logging.WithLogger(context, logger))
		context.Writer.Header().Set(XRequestIdKey, requestId)
	}
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				c.AbortWithStatus(http.StatusGatewayTimeout)
			}
			cancel()
		}()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func LoggingMiddleware(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skip[path] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, ok := skip[c.FullPath()]; ok {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		logger := logging.FromContext(c.Request.Context())
		timestamp := time.Now()
		latency := timestamp.Sub(start)
		latencyValue := latency.String()
		clientIP := c.ClientIP()
		method := c.Request.Method
		status := c.Writer.Status()
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}
		// append logger keys if not success or too slow latency.
		if status != http.StatusOK {
			logger = logger.With("status", status)
		}
		if latency > time.Second*3 {
			logger = logger.With("latency", latencyValue)
		}
		logger.Infof("[ARTICLE_API] %v | %3d | %s | %13v | %15s | %-7s %#v",
			timestamp.Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			latencyValue,
			clientIP,
			method,
			path,
		)
	}
}

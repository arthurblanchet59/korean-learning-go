package logging

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type AppLogger struct {
	accessFile *os.File
	errorFile  *os.File
	access     *log.Logger
	errors     *log.Logger
}

func New(logDir string) (*AppLogger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	accessFile, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	errorFile, err := os.OpenFile(filepath.Join(logDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		_ = accessFile.Close()
		return nil, err
	}

	return &AppLogger{
		accessFile: accessFile,
		errorFile:  errorFile,
		access:     log.New(accessFile, "INFO ", log.Ldate|log.Ltime|log.LUTC),
		errors:     log.New(errorFile, "ERROR ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile),
	}, nil
}

func (logger *AppLogger) Close() error {
	accessErr := logger.accessFile.Close()
	errorErr := logger.errorFile.Close()
	if accessErr != nil {
		return accessErr
	}
	return errorErr
}

func (logger *AppLogger) AccessMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()
		ctx.Next()

		latency := time.Since(startedAt)
		status := ctx.Writer.Status()
		entry := fmt.Sprintf(
			"%s %s status=%d latency=%s ip=%s userAgent=%q",
			ctx.Request.Method,
			ctx.Request.URL.RequestURI(),
			status,
			latency,
			ctx.ClientIP(),
			ctx.Request.UserAgent(),
		)

		logger.access.Println(entry)
		if status >= http.StatusBadRequest {
			logger.errors.Println(entry)
		}
		for _, err := range ctx.Errors {
			logger.errors.Printf("%s %s error=%q", ctx.Request.Method, ctx.Request.URL.RequestURI(), err.Error())
		}
	}
}

func (logger *AppLogger) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(ctx *gin.Context, recovered any) {
		logger.errors.Printf("panic method=%s path=%s recovered=%v", ctx.Request.Method, ctx.Request.URL.RequestURI(), recovered)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}

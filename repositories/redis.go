package repositories

import (
	"github.com/gin-gonic/gin"
)

type RedisRepo interface {
	Ping(ctx *gin.Context) error
	Get(ctx *gin.Context, key string) (string, error)
	Set(ctx *gin.Context, key, value string) error
	Delete(ctx *gin.Context, key string) error
	Keys(ctx *gin.Context, pattern string) ([]string, error)
}

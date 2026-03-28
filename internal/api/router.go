package api

import (
	v1 "ValorantAPI/internal/api/v1"
	"ValorantAPI/internal/deps"

	"github.com/gin-gonic/gin"
)

func LoadDefaultRouter(
	r *gin.Engine,
	deps *deps.Deps,
) {
	v1Group := r.Group("/v1")
	v1.RegisterRoutes(v1Group, deps)
}

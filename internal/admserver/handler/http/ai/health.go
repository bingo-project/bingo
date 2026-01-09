// ABOUTME: HTTP handler for AI Provider health monitoring in admin panel.
// ABOUTME: Provides endpoint to check health status of all AI providers.
package ai

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/pkg/core"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type HealthHandler struct {
	registry any
}

func NewHealthHandler(registry any) *HealthHandler {
	return &HealthHandler{registry: registry}
}

// listProviderRegistry is the interface we need from the registry.
type listProviderRegistry interface {
	ListProviders() []string
}

// GetHealthStatus
// @Summary    Get AI provider health status
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Success    200  {object}  v1.ListAiProviderHealthResponse
// @Failure    503  {object}  core.ErrResponse
// @Router     /v1/ai/health [GET].
func (h *HealthHandler) GetHealthStatus(c *gin.Context) {
	var providers []string

	if registry, ok := h.registry.(listProviderRegistry); ok {
		providers = registry.ListProviders()
	}

	data := make([]v1.AiProviderHealthInfo, 0, len(providers))

	for _, providerName := range providers {
		info := v1.AiProviderHealthInfo{
			ProviderName: providerName,
			Status:       "unknown",
			LastCheck:    time.Now(),
		}
		data = append(data, info)
	}

	core.Response(c, &v1.ListAiProviderHealthResponse{
		Total: int64(len(data)),
		Data:  data,
	}, nil)
}

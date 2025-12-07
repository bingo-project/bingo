package auth

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/log"

	_ "bingo/pkg/api/apiserver/v1"
)

// Providers
// @Summary	    Get auth providers
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Success	    200		{object}	[]v1.AuthProviderBrief
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/providers [GET].
func (ctrl *AuthController) Providers(c *gin.Context) {
	log.C(c).Infow("Providers function called")

	resp, err := ctrl.b.AuthProviders().FindEnabled(c)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

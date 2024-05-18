package auth

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	_ "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/core"
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
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

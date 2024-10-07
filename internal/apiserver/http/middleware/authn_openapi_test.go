package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bingo/internal/apiserver/config"
	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/model"

	"github.com/bingo-project/component-base/web/signer"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthnOpenAPI(t *testing.T) {
	Convey("TestAuthnOpenAPI", t, func() {
		facade.Config.OpenAPI = config.OpenAPI{
			Enabled: true,
			Nonce:   true,
			TTL:     60,
		}

		// Params
		params := map[string]any{
			"timestamp": time.Now().Unix(),
			"nonce":     uuid.New().String(),
		}

		// Request
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		body, _ := convertor.ToBytes(params)
		ctx.Request, _ = http.NewRequest("GET", "/path/to/resource", bytes.NewBuffer(body))

		Convey("disabled", func() {
			facade.Config.OpenAPI.Enabled = false

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeFalse)
		})

		Convey("error json", func() {
			ctx.Request, _ = http.NewRequest("GET", "/path/to/resource", nil)

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("error timestamp", func() {
			params["timestamp"] = time.Now().Unix() - 100
			body, _ = convertor.ToBytes(params)
			ctx.Request, _ = http.NewRequest("GET", "/path/to/resource", bytes.NewBuffer(body))

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("error nonce", func() {
			params["nonce"] = ""
			body, _ = convertor.ToBytes(params)
			ctx.Request, _ = http.NewRequest("GET", "/path/to/resource", bytes.NewBuffer(body))

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		// mock apiKey
		apiKey := &model.ApiKey{
			AccessKey: "xxx",
			SecretKey: "test-secret-key",
			Status:    model.ApiKeyStatusEnabled,
			ACL:       []string{},
			ExpiredAt: pointer.Of(carbon.Now().AddHour().StdTime()),
		}

		Convey("api key not found", func() {
			patches := gomonkey.ApplyPrivateMethod(store.S.ApiKeys(), "GetByAK", func(ctx context.Context, ak string) (*model.ApiKey, error) {
				return nil, errno.ErrResourceNotFound
			})
			defer patches.Reset()

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("api key disabled", func() {
			patches := gomonkey.ApplyPrivateMethod(store.S.ApiKeys(), "GetByAK", func(ctx context.Context, ak string) (*model.ApiKey, error) {
				apiKey.Status = model.ApiKeyStatusDisabled

				return apiKey, nil
			})
			defer patches.Reset()

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("api key expired", func() {
			patches := gomonkey.ApplyPrivateMethod(store.S.ApiKeys(), "GetByAK", func(ctx context.Context, ak string) (*model.ApiKey, error) {
				expiredAt := carbon.Now().SubHour().StdTime()
				apiKey.ExpiredAt = &expiredAt

				return apiKey, nil
			})
			defer patches.Reset()

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("ip not in acl", func() {
			patches := gomonkey.ApplyPrivateMethod(store.S.ApiKeys(), "GetByAK", func(ctx context.Context, ak string) (*model.ApiKey, error) {
				apiKey.ACL = []string{"100.100.100.100"}

				return apiKey, nil
			})
			defer patches.Reset()

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("error signature", func() {
			patches := gomonkey.ApplyPrivateMethod(store.S.ApiKeys(), "GetByAK", func(ctx context.Context, ak string) (*model.ApiKey, error) {
				return apiKey, nil
			})
			defer patches.Reset()

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeTrue)
		})

		Convey("pass", func() {
			// Sign
			client := signer.New()
			params["sign"] = client.Sign(params, apiKey.SecretKey)
			body, _ = convertor.ToBytes(params)
			ctx.Request, _ = http.NewRequest("GET", "/path/to/resource", bytes.NewBuffer(body))

			patches := gomonkey.ApplyPrivateMethod(store.S.ApiKeys(), "GetByAK", func(ctx context.Context, ak string) (*model.ApiKey, error) {
				return apiKey, nil
			})
			defer patches.Reset()

			AuthnOpenAPI()(ctx)
			So(ctx.IsAborted(), ShouldBeFalse)
		})
	})
}

package router

import (
	"context"
	"strings"

	"github.com/bingo-project/component-base/log"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/model"
	"bingo/internal/pkg/store"
)

func InitSystemAPI(g *gin.Engine) {
	// Get all routes
	routes := g.Routes()

	// Init api
	data := make([]model.ApiM, 0)
	for _, route := range routes {
		// Only system api
		if !strings.Contains(route.Path, "/v1") {
			continue
		}

		api := model.ApiM{
			Method: route.Method,
			Path:   route.Path,
			Group:  getGroup(route.Path),
		}

		data = append(data, api)
	}

	// Sort by path
	_ = slice.SortByField(data, "Path")

	for _, item := range data {
		// Create API.
		where := &model.ApiM{Method: item.Method, Path: item.Path}
		err := store.S.SysApi().FirstOrCreate(context.Background(), where, &item)
		if err != nil {
			log.Debugw("InitSystemAPI error", "err", err)

			break
		}
	}
}

func getGroup(path string) string {
	path = strings.TrimLeft(path, "/")
	pathArr := strings.Split(path, "/")

	// group
	group := ""
	if len(pathArr) > 1 {
		group = pathArr[1]
	}

	return group
}

package router

import (
	"context"
	"sort"
	"strings"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/model"
)

func InitAPI(g *gin.Engine) {
	// Get all routes
	routes := g.Routes()

	// Init api
	var data model.Apis
	for _, route := range routes {
		api := model.ApiM{
			Method: route.Method,
			Path:   route.Path,
			Group:  getGroup(route.Path),
		}

		data = append(data, api)
	}

	// Sort by path
	sort.Sort(data)

	for _, item := range data {
		// Create API.
		where := &model.ApiM{Method: item.Method, Path: item.Path}
		err := store.S.Api().FirstOrCreate(context.Background(), where, &item)
		if err != nil {
			log.Debugw("InitAPI error", "err", err)

			break
		}
	}
}

func getGroup(path string) string {
	pathArr := strings.Split(path, "/")

	// group
	group := pathArr[1]
	if len(pathArr) > 2 {
		group = pathArr[2]
	}
	if len(pathArr) > 3 && !strings.Contains(pathArr[3], ":") {
		group = group + "." + pathArr[3]
	}

	return group
}

// ABOUTME: Seeds core API records for menu association.
// ABOUTME: Creates APIs that need to be linked to menus, with internal flag support.

package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// coreAPIs defines APIs that can be associated with menus.
var coreAPIs = []model.ApiM{
	// Admin management
	{Method: "GET", Path: "/v1/admins", Group: "Admin", Description: "List admins"},
	{Method: "POST", Path: "/v1/admins", Group: "Admin", Description: "Create admin"},
	{Method: "GET", Path: "/v1/admins/:name", Group: "Admin", Description: "Get admin"},
	{Method: "PUT", Path: "/v1/admins/:name", Group: "Admin", Description: "Update admin"},
	{Method: "DELETE", Path: "/v1/admins/:name", Group: "Admin", Description: "Delete admin"},
	{Method: "PUT", Path: "/v1/admins/:name/change-password", Group: "Admin", Description: "Change admin password"},
	{Method: "PUT", Path: "/v1/admins/:name/roles", Group: "Admin", Description: "Set admin roles"},

	// Role management
	{Method: "GET", Path: "/v1/roles", Group: "Role", Description: "List roles"},
	{Method: "POST", Path: "/v1/roles", Group: "Role", Description: "Create role"},
	{Method: "GET", Path: "/v1/roles/all", Group: "Role", Description: "Get all roles"},
	{Method: "GET", Path: "/v1/roles/:name", Group: "Role", Description: "Get role"},
	{Method: "PUT", Path: "/v1/roles/:name", Group: "Role", Description: "Update role"},
	{Method: "DELETE", Path: "/v1/roles/:name", Group: "Role", Description: "Delete role"},
	{Method: "PUT", Path: "/v1/roles/:name/menus", Group: "Role", Description: "Set role menus"},
	{Method: "GET", Path: "/v1/roles/:name/menus", Group: "Role", Description: "Get role menus"},
	{Method: "PUT", Path: "/v1/roles/:name/apis", Group: "Role", Description: "Set role APIs", Internal: true},
	{Method: "GET", Path: "/v1/roles/:name/apis", Group: "Role", Description: "Get role APIs"},

	// Menu management
	{Method: "GET", Path: "/v1/menus", Group: "Menu", Description: "List menus"},
	{Method: "POST", Path: "/v1/menus", Group: "Menu", Description: "Create menu"},
	{Method: "GET", Path: "/v1/menus/tree", Group: "Menu", Description: "Get menu tree"},
	{Method: "GET", Path: "/v1/menus/:id", Group: "Menu", Description: "Get menu"},
	{Method: "PUT", Path: "/v1/menus/:id", Group: "Menu", Description: "Update menu"},
	{Method: "DELETE", Path: "/v1/menus/:id", Group: "Menu", Description: "Delete menu"},
	{Method: "POST", Path: "/v1/menus/:id/toggle-hidden", Group: "Menu", Description: "Toggle menu hidden"},

	// API management
	{Method: "GET", Path: "/v1/apis", Group: "Api", Description: "List APIs"},
	{Method: "POST", Path: "/v1/apis", Group: "Api", Description: "Create API"},
	{Method: "GET", Path: "/v1/apis/all", Group: "Api", Description: "Get all APIs"},
	{Method: "GET", Path: "/v1/apis/tree", Group: "Api", Description: "Get API tree"},
	{Method: "GET", Path: "/v1/apis/:id", Group: "Api", Description: "Get API"},
	{Method: "PUT", Path: "/v1/apis/:id", Group: "Api", Description: "Update API"},
	{Method: "DELETE", Path: "/v1/apis/:id", Group: "Api", Description: "Delete API"},

	// User management
	{Method: "GET", Path: "/v1/users", Group: "User", Description: "List users"},
	{Method: "POST", Path: "/v1/users", Group: "User", Description: "Create user"},
	{Method: "GET", Path: "/v1/users/:name", Group: "User", Description: "Get user"},
	{Method: "PUT", Path: "/v1/users/:name", Group: "User", Description: "Update user"},
	{Method: "DELETE", Path: "/v1/users/:name", Group: "User", Description: "Delete user"},
	{Method: "PUT", Path: "/v1/users/:name/change-password", Group: "User", Description: "Change user password"},

	// App management
	{Method: "GET", Path: "/v1/apps", Group: "App", Description: "List apps"},
	{Method: "POST", Path: "/v1/apps", Group: "App", Description: "Create app"},
	{Method: "GET", Path: "/v1/apps/:appid", Group: "App", Description: "Get app"},
	{Method: "PUT", Path: "/v1/apps/:appid", Group: "App", Description: "Update app"},
	{Method: "DELETE", Path: "/v1/apps/:appid", Group: "App", Description: "Delete app"},

	// API Key management
	{Method: "GET", Path: "/v1/api-keys", Group: "ApiKey", Description: "List API keys"},
	{Method: "POST", Path: "/v1/api-keys", Group: "ApiKey", Description: "Create API key"},
	{Method: "GET", Path: "/v1/api-keys/:id", Group: "ApiKey", Description: "Get API key"},
	{Method: "PUT", Path: "/v1/api-keys/:id", Group: "ApiKey", Description: "Update API key"},
	{Method: "DELETE", Path: "/v1/api-keys/:id", Group: "ApiKey", Description: "Delete API key"},

	// Config app management
	{Method: "GET", Path: "/v1/cfg/apps", Group: "CfgApp", Description: "List config apps"},
	{Method: "POST", Path: "/v1/cfg/apps", Group: "CfgApp", Description: "Create config app"},
	{Method: "GET", Path: "/v1/cfg/apps/:id", Group: "CfgApp", Description: "Get config app"},
	{Method: "PUT", Path: "/v1/cfg/apps/:id", Group: "CfgApp", Description: "Update config app"},
	{Method: "DELETE", Path: "/v1/cfg/apps/:id", Group: "CfgApp", Description: "Delete config app"},

	// Config management
	{Method: "GET", Path: "/v1/cfg/configs", Group: "Config", Description: "List configs"},
	{Method: "POST", Path: "/v1/cfg/configs", Group: "Config", Description: "Create config"},
	{Method: "GET", Path: "/v1/cfg/configs/:id", Group: "Config", Description: "Get config"},
	{Method: "PUT", Path: "/v1/cfg/configs/:id", Group: "Config", Description: "Update config"},
	{Method: "DELETE", Path: "/v1/cfg/configs/:id", Group: "Config", Description: "Delete config"},

	// File management
	{Method: "POST", Path: "/v1/file/upload", Group: "File", Description: "Upload file"},

	// AI Role management
	{Method: "GET", Path: "/v1/ai/roles", Group: "AI", Description: "List AI roles"},
	{Method: "POST", Path: "/v1/ai/roles", Group: "AI", Description: "Create AI role"},
	{Method: "GET", Path: "/v1/ai/roles/:id", Group: "AI", Description: "Get AI role"},
	{Method: "PUT", Path: "/v1/ai/roles/:id", Group: "AI", Description: "Update AI role"},
	{Method: "DELETE", Path: "/v1/ai/roles/:id", Group: "AI", Description: "Delete AI role"},

	// AI Provider management
	{Method: "GET", Path: "/v1/ai/providers", Group: "AI", Description: "List AI providers"},
	{Method: "GET", Path: "/v1/ai/providers/:id", Group: "AI", Description: "Get AI provider"},
	{Method: "PUT", Path: "/v1/ai/providers/:id", Group: "AI", Description: "Update AI provider"},

	// AI Model management
	{Method: "GET", Path: "/v1/ai/models", Group: "AI", Description: "List AI models"},
	{Method: "GET", Path: "/v1/ai/models/:id", Group: "AI", Description: "Get AI model"},
	{Method: "PUT", Path: "/v1/ai/models/:id", Group: "AI", Description: "Update AI model"},

	// AI Quota management
	{Method: "GET", Path: "/v1/ai/quotas", Group: "AI", Description: "List AI user quotas"},
	{Method: "GET", Path: "/v1/ai/quotas/:uid", Group: "AI", Description: "Get AI user quota"},
	{Method: "PUT", Path: "/v1/ai/quotas/:uid", Group: "AI", Description: "Update AI user quota"},
	{Method: "POST", Path: "/v1/ai/quotas/:uid/reset-daily", Group: "AI", Description: "Reset AI user daily tokens"},

	// AI Health monitoring
	{Method: "GET", Path: "/v1/ai/health", Group: "AI", Description: "Get AI provider health"},
}

type ApiSeeder struct{}

func (ApiSeeder) Signature() string {
	return "ApiSeeder"
}

func (ApiSeeder) Run() error {
	ctx := context.Background()

	for _, api := range coreAPIs {
		where := &model.ApiM{Method: api.Method, Path: api.Path}
		if err := store.S.SysApi().FirstOrCreate(ctx, where, &api); err != nil {
			return err
		}
	}

	return nil
}

// ABOUTME: Seeds system menus with API associations.
// ABOUTME: Creates menu tree structure and links menus to APIs via Method:Path references.

package seeder

import (
	"context"
	"strings"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type menuSeedData struct {
	ParentPath string // Used to find parent menu
	Title      string
	Name       string
	Path       string
	Icon       string
	Component  string
	Redirect   string
	Sort       int
	Type       string
	AuthCode   string
	ApiRefs    []string // Format: "METHOD:PATH"
}

var defaultMenus = []menuSeedData{
	// Dashboard
	{Title: "page.dashboard.title", Name: "Dashboard", Path: "/dashboard", Icon: "lucide:layout-dashboard", Redirect: "/analytics", Sort: -1, Type: "catalog"},
	{ParentPath: "/dashboard", Title: "page.dashboard.analytics", Name: "Analytics", Path: "/analytics", Icon: "lucide:area-chart", Component: "/dashboard/analytics/index", Sort: 1, Type: "menu"},
	{ParentPath: "/dashboard", Title: "page.dashboard.workspace", Name: "Workspace", Path: "/workspace", Icon: "carbon:workspace", Component: "/dashboard/workspace/index", Sort: 2, Type: "menu"},

	// System management
	{Title: "system.title", Name: "System", Path: "/system", Icon: "carbon:settings", Sort: 9997, Type: "catalog"},

	// Menu management
	{ParentPath: "/system", Title: "system.menu.title", Name: "SystemMenu", Path: "/system/menu", Icon: "carbon:menu", Component: "/system/menu/list", Sort: 1, Type: "menu", AuthCode: "System:Menu:List", ApiRefs: []string{"GET:/v1/menus"}},
	{ParentPath: "/system/menu", Title: "common.create", Name: "SystemMenuCreate", Type: "button", AuthCode: "System:Menu:Create", ApiRefs: []string{"POST:/v1/menus"}},
	{ParentPath: "/system/menu", Title: "common.edit", Name: "SystemMenuEdit", Type: "button", AuthCode: "System:Menu:Edit", ApiRefs: []string{"PUT:/v1/menus/:id"}},
	{ParentPath: "/system/menu", Title: "common.delete", Name: "SystemMenuDelete", Type: "button", AuthCode: "System:Menu:Delete", ApiRefs: []string{"DELETE:/v1/menus/:id"}},

	// Role management
	{ParentPath: "/system", Title: "system.role.title", Name: "SystemRole", Path: "/system/role", Icon: "mdi:account-group", Component: "/system/role/list", Sort: 2, Type: "menu", AuthCode: "System:Role:List", ApiRefs: []string{"GET:/v1/roles"}},
	{ParentPath: "/system/role", Title: "common.create", Name: "SystemRoleCreate", Type: "button", AuthCode: "System:Role:Create", ApiRefs: []string{"POST:/v1/roles"}},
	{ParentPath: "/system/role", Title: "common.edit", Name: "SystemRoleEdit", Type: "button", AuthCode: "System:Role:Edit", ApiRefs: []string{"PUT:/v1/roles/:name"}},
	{ParentPath: "/system/role", Title: "common.delete", Name: "SystemRoleDelete", Type: "button", AuthCode: "System:Role:Delete", ApiRefs: []string{"DELETE:/v1/roles/:name"}},
	{ParentPath: "/system/role", Title: "system.role.setMenus", Name: "SystemRoleSetMenus", Type: "button", AuthCode: "System:Role:SetMenus", ApiRefs: []string{"PUT:/v1/roles/:name/menus", "GET:/v1/roles/:name/menus"}},

	// Admin management
	{ParentPath: "/system", Title: "system.admin.title", Name: "SystemAdmin", Path: "/system/admin", Icon: "carbon:user-admin", Component: "/system/admin/list", Sort: 3, Type: "menu", AuthCode: "System:Admin:List", ApiRefs: []string{"GET:/v1/admins"}},
	{ParentPath: "/system/admin", Title: "common.create", Name: "SystemAdminCreate", Type: "button", AuthCode: "System:Admin:Create", ApiRefs: []string{"POST:/v1/admins"}},
	{ParentPath: "/system/admin", Title: "common.edit", Name: "SystemAdminEdit", Type: "button", AuthCode: "System:Admin:Edit", ApiRefs: []string{"PUT:/v1/admins/:name"}},
	{ParentPath: "/system/admin", Title: "common.delete", Name: "SystemAdminDelete", Type: "button", AuthCode: "System:Admin:Delete", ApiRefs: []string{"DELETE:/v1/admins/:name"}},

	// About
	{Title: "demos.vben.about", Name: "About", Path: "/about", Icon: "lucide:copyright", Component: "/_core/about/index", Sort: 9999, Type: "menu"},
}

type MenuSeeder struct{}

func (MenuSeeder) Signature() string {
	return "MenuSeeder"
}

func (MenuSeeder) Run() error {
	ctx := context.Background()

	// Build path -> menu map for parent lookup
	pathToMenu := make(map[string]*model.MenuM)

	for _, data := range defaultMenus {
		menu := &model.MenuM{
			Title:     data.Title,
			Name:      data.Name,
			Path:      data.Path,
			Icon:      data.Icon,
			Component: data.Component,
			Redirect:  data.Redirect,
			Sort:      data.Sort,
			Type:      data.Type,
			AuthCode:  data.AuthCode,
			Status:    "enabled",
			Hidden:    data.Type == "button",
		}

		// Set parent ID
		if data.ParentPath != "" {
			if parent, ok := pathToMenu[data.ParentPath]; ok {
				menu.ParentID = parent.ID
			}
		}

		// Find and associate APIs
		if len(data.ApiRefs) > 0 {
			var apis []*model.ApiM
			for _, ref := range data.ApiRefs {
				parts := strings.SplitN(ref, ":", 2)
				if len(parts) != 2 {
					continue
				}
				api, err := store.S.SysApi().GetByMethodPath(ctx, parts[0], parts[1])
				if err != nil {
					continue
				}
				apis = append(apis, api)
			}
			menu.Apis = apis
		}

		// Create menu with APIs (idempotent)
		if err := store.S.SysMenu().FirstOrCreateWithApis(ctx, menu); err != nil {
			return err
		}

		// Store in map for parent lookup
		if menu.Path != "" {
			pathToMenu[menu.Path] = menu
		}
	}

	return nil
}

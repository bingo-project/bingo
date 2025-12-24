// ABOUTME: Seeds role-menu associations.
// ABOUTME: Assigns menus to roles and syncs API permissions to Casbin.

package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type roleMenuAssignment struct {
	RoleName  string
	MenuPaths []string // Menu paths to assign
	AllMenus  bool     // Assign all menus
}

var roleMenuAssignments = []roleMenuAssignment{
	{
		RoleName: "super-admin",
		AllMenus: true,
	},
	{
		RoleName: "admin",
		MenuPaths: []string{
			"/system",
			"/system/admin",
			"/system/role",
			"/system/menu",
		},
	},
}

type RoleMenuSeeder struct{}

func (RoleMenuSeeder) Signature() string {
	return "RoleMenuSeeder"
}

func (RoleMenuSeeder) Run() error {
	ctx := context.Background()

	// Get all menus
	allMenus, err := store.S.SysMenu().All(ctx)
	if err != nil {
		return err
	}

	// Build path -> menu map
	pathToMenu := make(map[string]uint)
	for _, menu := range allMenus {
		if menu.Path != "" {
			pathToMenu[menu.Path] = menu.ID
		}
	}

	authz, err := auth.NewAuthorizer(store.S.DB(ctx), nil)
	if err != nil {
		return err
	}

	for _, assignment := range roleMenuAssignments {
		// Get role
		role, err := store.S.SysRole().GetByName(ctx, assignment.RoleName)
		if err != nil {
			continue
		}

		var menus []*model.MenuM

		if assignment.AllMenus {
			// Assign all menus with APIs
			menus, err = store.S.SysMenu().AllWithApis(ctx)
			if err != nil {
				return err
			}
		} else {
			// Get menu IDs from paths
			var menuIDs []uint
			for _, path := range assignment.MenuPaths {
				if id, ok := pathToMenu[path]; ok {
					menuIDs = append(menuIDs, id)
				}
			}

			if len(menuIDs) == 0 {
				continue
			}

			// Get menus with APIs
			menus, err = store.S.SysMenu().GetByIDsWithApis(ctx, menuIDs)
			if err != nil {
				return err
			}
		}

		// Update role menus
		role.Menus = menus
		if err := store.S.SysRole().UpdateWithMenus(ctx, role); err != nil {
			return err
		}

		// Extract API IDs and sync to Casbin
		apiIDs := make(map[uint]struct{})
		for _, menu := range menus {
			for _, api := range menu.Apis {
				apiIDs[api.ID] = struct{}{}
			}
		}

		ids := make([]uint, 0, len(apiIDs))
		for id := range apiIDs {
			ids = append(ids, id)
		}

		// Get APIs and add to Casbin
		if len(ids) > 0 {
			apis, err := store.S.SysApi().GetByIDs(ctx, ids)
			if err != nil {
				return err
			}

			rules := make([][]string, 0, len(apis))
			for _, api := range apis {
				rules = append(rules, []string{known.RolePrefix + role.Name, api.Path, api.Method})
			}

			if _, err := authz.Enforcer().AddPolicies(rules); err != nil {
				return err
			}
		}
	}

	return nil
}

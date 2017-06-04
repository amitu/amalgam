package db

import (
	"context"
	"fmt"

	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
	"github.com/juju/errors"
)

type group struct {
	DId   int64  `db:"id"`
	DName string `db:"name"`

	store *store
}

func (g *group) ID() int64 {
	return g.DId
}

func (g *group) Name() string {
	return g.DName
}

func (g *group) Permissions(ctx context.Context) ([]django.Permission, error) {
	perms := []*permission{}
	q := `
		SELECT
			auth_permission.id,
			auth_permission.name,
			auth_permission.codename
		FROM
			auth_permission
		JOIN auth_group_permissions
			ON
				(auth_permission.id = auth_group_permissions.permission_id)
		WHERE
			group_id = $1
	`

	err := amalgam.QueryIntoSlice(ctx, &perms, q, g.DId)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cperms := []django.Permission{}
	for _, g := range perms {
		cperms = append(cperms, g)
	}

	return cperms, nil
}

type permission struct {
	DId   int64  `db:"id"`
	DName string `db:"name"`
	DCode string `db:"codename"`

	store *astore
}

func (p *permission) ID() int64 {
	return p.DId
}

func (p *permission) Name() string {
	return p.DName
}

func (p *permission) Code() string {
	return p.DCode
}

type user struct {
	DID        int64  `db:"id" json:"id"`
	DUsername  string `db:"username" json:"-"`
	DFirstName string `db:"first_name" json:"first_name"`
	DLastName  string `db:"last_name" json:"last_name"`

	store *astore
}

func (u *user) ID() int64 {
	return u.DID
}

func (u *user) Name() string {
	return u.DFirstName + " " + u.DLastName
}

func (u *user) Email() string {
	panic("not implemented")
	return ""
}

func (u *user) CheckPassword(string) bool {
	panic("not implemented")
	return false
}

func (u *user) Roles() ([]string, error) {
	panic("not implemented")
	return nil, nil
}

func (u *user) Permissions() ([]django.Permission, error) {
	panic("not implemented")
	return nil, nil
}

func (u *user) HasRole(ctx context.Context, roleId int64) (bool, error) {
	query := "SELECT count(*) FROM " + u.store.UserGroupsTable +
		" WHERE user_id = $1 AND group_id = $2"
	num, err := amalgam.QueryIntoInt(ctx, query, u.DID, roleId)
	if err != nil {
		return false, errors.Trace(err)
	}

	return num != 0, nil
}

func (u *user) HasPermission(string) (bool, error) {
	panic("not implemented")
	return false, nil
}

func (u *user) SetName(string, bool) error {
	panic("not implemented")
	return nil
}

func (u *user) SetEmail(string, bool) error {
	panic("not implemented")
	return nil
}

func (u *user) SetPassword(string, bool) error {
	panic("not implemented")
	return nil
}

func (u *user) Save() error {
	panic("not implemented")
	return nil
}

func (u *user) Deactivate(reason string) error {
	panic("not implemented")
	return nil
}

func (u *user) IsSuperUser() bool {
	panic("not implemented")
	return false
}

type AuthTables struct {
	UserTable             string
	GroupTable            string
	PermissionTable       string
	UserGroupsTable       string
	UserPermissionsTable  string
	GroupPermissionsTable string
}

func updateAuthTablesWithDefault(at *AuthTables) {
	if at.UserTable == "" {
		at.UserTable = "auth_user"
	}
	if at.UserTable == "" {
		at.GroupTable = "auth_group"
	}
	if at.PermissionTable == "" {
		at.PermissionTable = "auth_permission"
	}
	if at.UserGroupsTable == "" {
		at.UserGroupsTable = "auth_user_groups"
	}
	if at.UserPermissionsTable == "" {
		at.UserPermissionsTable = "auth_user_user_permissions"
	}
	if at.GroupPermissionsTable == "" {
		at.GroupPermissionsTable = "auth_group_permissions"
	}
}

type astore struct {
	AuthTables
}

func (s *astore) Groups(ctx context.Context) ([]django.Group, error) {
	groups := []*group{}
	query := fmt.Sprintf("SELECT * FROM %s", s.GroupTable)
	err := amalgam.QueryIntoSlice(ctx, &groups, query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cgroups := []django.Group{}
	for _, g := range groups {
		cgroups = append(cgroups, g)
	}

	return cgroups, nil
}

func (s *astore) GroupByID(ctx context.Context, id int64) (django.Group, error) {
	group := group{}
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", s.GroupTable)
	err := amalgam.QueryIntoStruct(ctx, &group, query, id)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &group, nil
}

func (s *astore) GroupByName(
	ctx context.Context, name string,
) (django.Group, error) {
	group := group{}
	query := fmt.Sprintf("SELECT * FROM %s WHERE name = $1", s.GroupTable)
	err := amalgam.QueryIntoStruct(ctx, &group, query, name)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &group, nil
}

func (s *astore) UserByID(ctx context.Context, id int64) (django.User, error) {
	u := &user{}
	query := "SELECT id, username, first_name, last_name FROM " + s.UserTable +
		" WHERE id = $1"
	//query := fmt.Sprint(`
	//	SELECT
	//		id, username, first_name, last_name
	//	FROM
	//		%s
	//	WHERE
	//		id = $1
	//`, s.UserTable,
	//)
	err := amalgam.QueryIntoStruct(ctx, u, query, id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	u.store = s
	return u, nil
}

func (s *astore) UserByEmail(context.Context, string) (django.User, error) {
	panic("not implemented")
	return nil, nil
}

func (s *astore) Authenticate(
	context.Context, string, string,
) (django.User, error) {
	panic("not implemented")
	return nil, nil
}

func (s *astore) Permissions(ctx context.Context) ([]django.Permission, error) {
	panic("not implemented")
	return nil, nil
}

func (s *astore) PermissionByID(ctx context.Context, id int64) (django.Permission, error) {
	perm := permission{}
	query := fmt.Sprintf("select * from %s where id = $1", s.PermissionTable)
	err := amalgam.QueryIntoStruct(ctx, &permission{}, query, id)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &perm, nil
}

func (s *astore) PermissionByCode(
	ctx context.Context, code string,
) (django.Permission, error) {
	perm := permission{}
	query := fmt.Sprintf("select * from %s where codename = $1", s.PermissionTable)
	err := amalgam.QueryIntoStruct(ctx, &permission{}, query, code)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &perm, nil
}

func NewCustomAuthStore(ctx context.Context, auth_tables AuthTables) (django.AuthStore, error) {
	updateAuthTablesWithDefault(&auth_tables)
	query := fmt.Sprintf("SELECT count(*) FROM %s", auth_tables.GroupTable)
	gcount, err := amalgam.QueryIntoInt(ctx, query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	query = fmt.Sprintf("SELECT count(*) FROM %s", auth_tables.PermissionTable)
	pcount, err := amalgam.QueryIntoInt(ctx, query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	query = fmt.Sprintf("SELECT count(*) FROM %s", auth_tables.UserTable)
	ucount, err := amalgam.QueryIntoInt(ctx, query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	amalgam.LOGGER.Debug(
		"found_auth_tables", "groups", gcount,
		"permissions", pcount, "users", ucount,
	)
	return &astore{auth_tables}, nil
}

func NewAuthStore(ctx context.Context) (django.AuthStore, error) {
	auth_tables := AuthTables{}
	updateAuthTablesWithDefault(&auth_tables)
	return NewCustomAuthStore(ctx, auth_tables)
}

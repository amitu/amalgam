package db

import (
	"context"

	amalgam "github.com/amitu/amalgam"
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
	q := `
		SELECT
			count(*)
		FROM
			auth_user_groups
		WHERE
			user_id = $1 AND
			group_id = $2
	`
	num, err := amalgam.QueryIntoInt(ctx, q, u.DID, roleId)
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

type astore struct {
}

func (s *astore) Groups(ctx context.Context) ([]django.Group, error) {
	groups := []*group{}
	err := amalgam.QueryIntoSlice(ctx, &groups, "SELECT * FROM auth_group")
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
	err := amalgam.QueryIntoStruct(
		ctx, &group, "SELECT * FROM auth_group WHERE id = $1", id,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &group, nil
}

func (s *astore) GroupByName(
	ctx context.Context, name string,
) (django.Group, error) {
	group := group{}
	err := amalgam.QueryIntoStruct(
		ctx, &group, `SELECT * FROM auth_group WHERE name = $1`, name,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &group, nil
}

func (s *astore) UserByID(ctx context.Context, id int64) (django.User, error) {
	u := &user{}
	err := amalgam.QueryIntoStruct(
		ctx, u, `
			SELECT
				id, username, first_name, last_name
			FROM
				auth_user
			WHERE
				id = $1
		`, id,
	)
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
	err := amalgam.QueryIntoStruct(
		ctx, &permission{}, "select * from auth_permission where id = $1", id,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &perm, nil
}

func (s *astore) PermissionByCode(
	ctx context.Context, code string,
) (django.Permission, error) {
	perm := permission{}
	err := amalgam.QueryIntoStruct(
		ctx, &permission{}, "select * from auth_permission where codename = $1", code,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &perm, nil
}

func NewAuthStore(ctx context.Context) (django.AuthStore, error) {
	gcount, err := amalgam.QueryIntoInt(ctx, "SELECT count(*) FROM auth_group")
	if err != nil {
		return nil, errors.Trace(err)
	}

	pcount, err := amalgam.QueryIntoInt(
		ctx, "SELECT count(*) FROM auth_permission",
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ucount, err := amalgam.QueryIntoInt(ctx, "SELECT count(*) FROM auth_user")
	if err != nil {
		return nil, errors.Trace(err)
	}

	amalgam.LOGGER.Debug(
		"found_auth_tables", "groups", gcount,
		"permissions", pcount, "users", ucount,
	)
	return &astore{}, nil
}

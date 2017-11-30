package db

import (
	"context"
	"fmt"
	"strconv"

	"database/sql"
	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
	"github.com/juju/errors"
	"time"
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
	DID     int64 `db:"id" json:"id"`
	DFields map[string]interface{}
	store   *astore
}

func (u *user) ID() int64 {
	return u.DID
}

func (u *user) Field(key string) (interface{}, bool) {
	val, ok := u.DFields[key]
	if !ok {
		return nil, false
	}

	return val, true
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

func (u *user) Save(ctx context.Context) error {
	var values string
	index := 1
	var fields []interface{}
	for k, v := range u.DFields {
		if k == "id" {
			continue
		}
		values = values + k + "=$" + strconv.Itoa(index) + ","
		fields = append(fields, v)
		index++
	}
	values = values + "id" + "=$" + strconv.Itoa(index)
	fields = append(fields, u.DFields["id"])

	query := "UPDATE " + u.store.UserTable + " SET " +
		values + "WHERE id=$" + strconv.Itoa(index) + ";"

	err := amalgam.Exec(ctx, query, fields...)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (u *user) RefreshFromDB(ctx context.Context) error {
	query := "SELECT * FROM " + u.store.AuthTables.UserTable +
		" WHERE id = $1"

	userMap, err := amalgam.QueryIntoMap(ctx, query, u.DID)
	if err != nil {
		return errors.Trace(err)
	}
	u.DFields = userMap

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
	query := "SELECT * FROM " + s.UserTable +
		" WHERE id = $1"

	userMap, err := amalgam.QueryIntoMap(ctx, query, id)
	if err != nil {
		return u, errors.Trace(err)
	}

	u.DID = userMap["id"].(int64)
	u.DFields = userMap
	u.store = s

	return u, nil
}

func (s *astore) UserByAPIKey(
	ctx context.Context, apiKey string,
) (django.User, error) {
	u := &user{}
	query := "SELECT * FROM " + s.UserTable +
		" WHERE id = (select user_id from acko_userprofile where api_key= $1)"

	userMap, err := amalgam.QueryIntoMap(ctx, query, apiKey)
	if err != nil {
		return u, errors.Trace(err)
	}

	u.DID = userMap["id"].(int64)
	u.DFields = userMap
	u.store = s

	return u, nil
}

func (s *astore) UserByPhone(
	ctx context.Context, phone string,
) (django.User, error) {
	u := &user{}
	query := "SELECT * FROM " + s.UserTable +
		" WHERE phone = $1"

	userMap, err := amalgam.QueryIntoMap(ctx, query, phone)
	if err != nil {
		return u, errors.Trace(err)
	}

	u.DID = userMap["id"].(int64)
	u.DFields = userMap
	u.store = s

	return u, nil
}

func (s *astore) GetOrCreateUser(
	ctx context.Context, details map[string]interface{},
) (django.User, error) {
	u := &user{}
	query := "SELECT * FROM " + s.UserTable +
		" WHERE phone = $1"

	phone := details["phone"].(string)
	first_name := details["first_name"].(string)
	last_name := details["first_name"].(string)

	userMap, err := amalgam.QueryIntoMap(ctx, query, phone)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			query := "INSERT INTO " +
				s.UserTable + "(" + "phone, password, is_superuser, " +
				"first_name, " + "last_name, is_staff, is_active, joined_on, " +
				"available, idle, is_online, on_call) " + "VALUES " +
				"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"

			err := amalgam.Exec(
				ctx, query, phone, "asd", false, first_name, last_name,
				false, false, time.Now(), false, true, false, false,
			)
			if err != nil {
				return nil, errors.Trace(err)
			}

			query = "SELECT * FROM " + s.UserTable +
				" WHERE phone = $1"

			userMap, err = amalgam.QueryIntoMap(ctx, query, phone)
		} else {
			return u, errors.Trace(err)
		}
	}

	u.DID = userMap["id"].(int64)
	u.DFields = userMap
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

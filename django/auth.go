package django

import "context"

type User interface {
	ID() int64
	Field(string) (interface{}, bool)
	Email() string
	CheckPassword(string) bool

	Roles() ([]string, error)
	Permissions() ([]Permission, error)
	HasRole(context.Context, int64) (bool, error)
	HasPermission(string) (bool, error)

	SetName(string, bool) error
	SetEmail(string, bool) error
	SetPassword(string, bool) error
	Save(context.Context) error
	RefreshFromDB(context.Context) error
	Deactivate(reason string) error

	IsSuperUser() bool
}

type UserStore interface {
	GetOrCreateUser(context.Context, map[string]interface{}) (User, error)
	UserByID(context.Context, int64) (User, error)
	UserByAPIKey(context.Context, string) (User, error)
	UserByPhone(context.Context, string) (User, error)
	UserByEmail(context.Context, string) (User, error)
	Authenticate(context.Context, string, string) (User, error)
}

type Group interface {
	ID() int64
	Name() string
	Permissions(context.Context) ([]Permission, error)
}

type GroupStore interface {
	Groups(context.Context) ([]Group, error)
	GroupByID(context.Context, int64) (Group, error)
	GroupByName(context.Context, string) (Group, error)
}

type Permission interface {
	ID() int64
	Code() string
	Name() string
}

type PermissionStore interface {
	Permissions(context.Context) ([]Permission, error)
	PermissionByID(context.Context, int64) (Permission, error)
	PermissionByCode(context.Context, string) (Permission, error)
}

type AuthStore interface {
	UserStore
	GroupStore
	PermissionStore
}

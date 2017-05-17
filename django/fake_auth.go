package django

import (
    "context"
)

func FakeUser() User {
	return &FUser{}
}

type FUser struct {
}

func (u *FUser) ID() int64 {
	panic("not yet implemented")
}

func (u *FUser) Name() string {
	panic("not yet implemented")
}

func (u *FUser) Email() string {
	panic("not yet implemented")
}

func (u *FUser) CheckPassword(str string) bool {
	panic("not yet implemented")
}

func (u *FUser) Roles() ([]string, error) {
	panic("not yet implemented")
}

func (u *FUser) Permissions() ([]Permission, error) {
	panic("not yet implemented")
}

func (u *FUser) HasRole(ctx context.Context, i int64) (bool, error) {
	return true, nil
}

func (u *FUser) HasPermission(str string) (bool, error) {
	panic("not yet implemented")
}

func (u *FUser) SetName(str string, b bool) error {
	panic("not yet implemented")
}

func (u *FUser) SetEmail(str string, b bool) error {
	panic("not yet implemented")
}

func (u *FUser) SetPassword(str string, b bool) error {
	panic("not yet implemented")
}

func (u *FUser) Save() error {
	panic("not yet implemented")
}

func (u *FUser) Deactivate(str string) error {
	panic("not yet implemented")
}

func (u *FUser) IsSuperUser() bool {
	panic("not yet implemented")
}

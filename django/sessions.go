package django

import "context"

const (
	KeyUserID    = "_auth_user_id"
	KeyCSRFToken = "_csrf_token"
)

type Session interface {
	// ID() returns the unique opaque if for the session. This id should be stored
	// in cookie etc.
	SessionKey() string
	// SetValue() stores a value in the session against the passed key. This will
	// be in store till the session is destroyed or till DeleteValue() is called.
	// Can return an error in case persistence fails. If this method is called
	// multiple times with same key, value keeps overwriting the old value.
	SetValue(ctx context.Context, key string, value interface{}) error
	// DeleteValue() removes the key from the session. If key is not present no
	// error is returned, only when there is an error during saving the session an
	// error is returned.
	//DeleteValue(key string) error
	// GetValue() returns a value, or ErrNotFound if there is no such value in the
	// session.
	GetValue(string) ([]byte, error)
	// GetInt() returns the value as Int, it will return ErrNotInt in case the
	// value stored is not an int.
	GetInt64(string) (int64, error)
	// GetString() returns the value as string, it will return ErrNotInt in case
	// the value stored is not a string.
	GetString(string) (string, error)
	GetUser(context.Context) (User, error)
	// Destroy() destroys the session from session store. Any calls to any method
	// session object after that may lead to error or crash.
	Destroy(context.Context) error

	Store() SessionStore

	String() string
}

type SessionStore interface {
	// If GetSession on store is called with id of a destroyed session, a fresh
	// session is created and returned.
	GetSessionBySessionKey(ctx context.Context, key string) (Session, error)
	CreateSession(context.Context) (Session, error)
	DestroySession(ctx context.Context, id string) error
}

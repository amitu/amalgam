package django

import (
	"context"
	"encoding/json"

	"acko"

	"github.com/juju/errors"
)

// NewSessionStore creates a session store. It varifies that the session related
// table exists (but not if the table schema is correct).
func NewFakeSessionStore() SessionStore {
	return &fakeSessionStore{make(map[string]*session)}
}

type session struct {
	store  SessionStore
	id     string
	values map[string]json.RawMessage
}

type fakeSessionStore struct {
	sessions map[string]*session
}

func (f *fakeSessionStore) GetSessionBySessionKey(
	_ context.Context, id string,
) (Session, error) {
	s, ok := f.sessions[id]
	if !ok {
		return nil, errors.New("no such session")
	}
	return s, nil
}

func (f *fakeSessionStore) CreateSession(_ context.Context) (Session, error) {
	ss := session{}
	ss.id = acko.GetRandomString(32)
	ss.store = f
	ss.values = make(map[string]json.RawMessage)

	f.sessions[ss.id] = &ss
	return &ss, nil
}

func (f *fakeSessionStore) DestroySession(_ context.Context, id string) error {
	panic("not implemented")
	return nil // TODO
}

func (s *session) SessionKey() string {
	return s.id
}

func (s *session) SetValue(ctx context.Context, key string, value interface{}) error {
	svalue, err := json.Marshal(value)
	if err != nil {
		return errors.Trace(err)
	}
	s.values[key] = svalue
	return nil
}

func (s *session) DeleteValue(key string) error {
	delete(s.values, key)
	return nil
}

func (s *session) GetValue(key string) ([]byte, error) {
	return s.values[key], nil
}

func (s *session) GetInt64(key string) (int64, error) {
	v, err := s.GetValue(key)

	if err != nil {
		return 0, errors.Trace(err)
	}

	i := int64(0)
	err = json.Unmarshal(v, &i)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return i, nil
}

func (s *session) GetString(key string) (string, error) {
	v, err := s.GetValue(key)

	if err != nil {
		return "", errors.Trace(err)
	}

	var svalue string
	err = json.Unmarshal(v, &svalue)
	if err != nil {
		return "", errors.Trace(err)
	}

	return svalue, nil
}

func (s *session) GetUser(context.Context) (User, error) {
	panic("not implemented")
	return nil, nil // TODO
}

func (s *session) Destroy(context.Context) error {
	panic("not implemented")
	return nil // TODO
}

func (s *session) Store() SessionStore {
	return s.store
}

func (s *session) String() string {
	panic("not implemented")
	return ""
}

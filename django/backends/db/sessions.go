package db

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	amalgam "github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
	"github.com/juju/errors"
)

var (
	insert_query = `
		INSERT INTO django_session
			(expire_date, session_data, session_key)
		VALUES
			($1, $2, $3)
	`
)

// NewSessionStore creates a session store. It verifies that the session related
// table exists (but not if the table schema is correct).
func NewSessionStore(
	ctx context.Context, secret string, auth django.AuthStore,
) (django.SessionStore, error) {
	count, err := amalgam.QueryIntoInt(
		ctx, "SELECT count(*) FROM django_session",
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	amalgam.LOGGER.Debug("found_session_table", "count", count)
	return &store{secret: secret, auth: auth}, nil
}

type store struct {
	secret string
	auth   django.AuthStore
}

func (s *store) GetSessionBySessionKey(
	ctx context.Context, id string,
) (django.Session, error) {
	ss := session{}
	ss.store = s
	ss.dDataCache = make(map[string]json.RawMessage)
	ss.DSessionKey = id

	err := amalgam.QueryIntoStruct(
		ctx, &ss, `SELECT * FROM django_session WHERE session_key = $1`, id,
	)
	if err != nil {
		return &ss, errors.Trace(err)
	}
	ss.dDataCache = nil

	return &ss, nil
}

func (s *store) CreateSession(ctx context.Context) (django.Session, error) {
	session_key := amalgam.GetRandomString(32)
	expire_date := time.Now().Add(time.Hour * 24 * 3600)

	ss := session{}
	ss.DSessionKey = session_key
	ss.DExpireDate = expire_date
	ss.store = s
	ss.dDataCache = make(map[string]json.RawMessage)

	err := ss.Save(ctx, true)
	if err != nil {
		return nil, errors.Trace(err)
	}

	amalgam.LOGGER.Debug("Session Saved: ", ss.SessionKey())

	return &ss, nil
}

func (s *store) DestroySession(ctx context.Context, id string) error {
	tx, err := amalgam.Ctx2Tx(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	_, err = tx.Exec("DELETE FROM django_session WHERE session_key=?", id)
	return errors.Trace(err)
}

type session struct {
	DSessionKey string    `db:"session_key"`
	DExpireDate time.Time `db:"expire_date"`
	DData       string    `db:"session_data"`

	store      *store
	dDataCache map[string]json.RawMessage
}

func (s *session) String() string {
	if s.dDataCache == nil {
		dDataCache, err := s.loadSessionData()
		if err != nil {
			return errors.ErrorStack(err)
		}
		s.dDataCache = dDataCache
	}

	return fmt.Sprintf(
		"id=%s data=%s cache=%v", s.DSessionKey, s.DData, s.dDataCache,
	)
}

func (s *session) hashData(data []byte) []byte {
	salt := "django.contrib.sessions.SessionStore"

	key := sha1.Sum([]byte(salt + s.store.secret))

	mac := hmac.New(sha1.New, key[:])
	mac.Write(data)

	return mac.Sum(nil)
}

func (s *session) checkHash(mac []byte, data []byte) bool {
	expectedMac := s.hashData(data)
	return hmac.Equal(mac, expectedMac)
}

func (s *session) loadSessionData() (map[string]json.RawMessage, error) {
	data, err := base64.StdEncoding.DecodeString(s.DData)
	if err != nil {
		return nil, errors.Trace(err)
	}
	split := strings.SplitN(string(data), ":", 2)
	hash := []byte(split[0])
	jdata := []byte(split[1])

	if !s.checkHash(hash, jdata) && false {
		return nil, errors.New("hash is incorrect won't parse session data")
	}

	m := make(map[string]json.RawMessage)
	err = json.Unmarshal([]byte(jdata), &m)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return m, nil
}

func (s *session) SessionKey() string {
	return s.DSessionKey
}

func (s *session) SetValue(
	ctx context.Context, key string, value interface{},
) error {
	svalue, err := json.Marshal(value)
	if err != nil {
		return errors.Trace(err)
	}
	amalgam.LOGGER.Debug("session_set_value", "value", svalue)
	if s.dDataCache == nil {
		s.dDataCache = make(map[string]json.RawMessage)
	}

	s.dDataCache[key] = json.RawMessage(svalue)

	return errors.Trace(s.Save(ctx, false))
}

func (s *session) DeleteValue(ctx context.Context, key string) error {
	delete(s.dDataCache, key)
	return errors.Trace(s.Save(ctx, false))
}

func (s *session) GetValue(key string) ([]byte, error) {
	if s.dDataCache == nil {
		dDataCache, err := s.loadSessionData()
		if err != nil {
			return nil, errors.Trace(err)
		}
		s.dDataCache = dDataCache
	}

	v, ok := s.dDataCache[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	return []byte(v), nil
}

func (s *session) GetInt64(key string) (int64, error) {
	val, err := s.GetValue(key)
	if err != nil {
		return 0, errors.Trace(err)
	}

	var intval int64
	err = json.Unmarshal(val, &intval)

	if err != nil {
		return 0, errors.Trace(err)
	}

	return intval, nil
}

func (s *session) GetString(key string) (string, error) {
	val, err := s.GetValue(key)
	if err != nil {
		return "", errors.Trace(err)
	}

	var strval string
	err = json.Unmarshal(val, &strval)

	if err != nil {
		return "", errors.Trace(err)
	}

	return strval, nil
}

func (s *session) Destroy(ctx context.Context) error {
	return s.store.DestroySession(ctx, s.DSessionKey)
}

func (s *session) Store() django.SessionStore {
	return s.store
}

func (s *session) GetUser(ctx context.Context) (django.User, error) {
	uid, err := s.GetInt64(django.KeyUserID)
	if err != nil {
		uid, err := s.GetString(django.KeyUserID)
		if err != nil {
			return nil, errors.Trace(err)
		}
		uid_int, err := strconv.ParseInt(uid, 0, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		user, err := s.store.auth.UserByID(ctx, uid_int)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return user, nil
	}
	user, err := s.store.auth.UserByID(ctx, uid)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return user, nil
}

func (s *session) prepareForSave() error {
	// serialize DData
	data, err := json.Marshal(s.dDataCache)
	if err != nil {
		return errors.Trace(err)
	}

	str := []byte(hex.EncodeToString(s.hashData(data)) + ":")
	str = append(str, data...)
	s.DData = base64.StdEncoding.EncodeToString(str)

	return nil
}

func (s *session) Save(ctx context.Context, i bool) error {
	err := s.prepareForSave()
	if err != nil {
		return errors.Trace(err)
	}

	tx, err := amalgam.Ctx2Tx(ctx)
	if err != nil {
		return errors.Trace(err)
	}

	result, err := tx.Exec(
		"UPDATE django_session SET session_data = $1 WHERE session_key = $2",
		s.DData, s.DSessionKey,
	)
	if err != nil {
		return errors.Trace(err)
	}

	num, err := result.RowsAffected()
	if err != nil {
		return errors.Trace(err)
	}

	if num == 0 {
		amalgam.LOGGER.Debug("db_insert_session", "sessionkey", s.DSessionKey)
		_, err = tx.Exec(insert_query, s.DExpireDate, s.DData, s.DSessionKey)
		if err != nil {
			return errors.Trace(err)
		}
	} else if num > 1 {
		return errors.New("Unexpected number of updates")
	}

	return nil
}

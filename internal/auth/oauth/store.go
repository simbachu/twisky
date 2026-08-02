package oauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"

	_ "modernc.org/sqlite"
)

// SQLiteStore persists indigo OAuth auth-request and session data.
type SQLiteStore struct {
	db *sql.DB
}

var _ indigooauth.ClientAuthStore = (*SQLiteStore)(nil)

func OpenSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS oauth_sessions (
	account_did TEXT NOT NULL,
	session_id TEXT NOT NULL,
	payload TEXT NOT NULL,
	PRIMARY KEY (account_did, session_id)
);
CREATE TABLE IF NOT EXISTS oauth_auth_requests (
	state TEXT PRIMARY KEY,
	payload TEXT NOT NULL
);
`)
	if err != nil {
		return fmt.Errorf("migrate oauth store: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetSession(ctx context.Context, did syntax.DID, sessionID string) (*indigooauth.ClientSessionData, error) {
	var payload string
	err := s.db.QueryRowContext(ctx,
		`SELECT payload FROM oauth_sessions WHERE account_did = ? AND session_id = ?`,
		did.String(), sessionID,
	).Scan(&payload)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", did)
	}
	if err != nil {
		return nil, err
	}
	var sess indigooauth.ClientSessionData
	if err := json.Unmarshal([]byte(payload), &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *SQLiteStore) SaveSession(ctx context.Context, sess indigooauth.ClientSessionData) error {
	payload, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO oauth_sessions (account_did, session_id, payload)
VALUES (?, ?, ?)
ON CONFLICT(account_did, session_id) DO UPDATE SET payload = excluded.payload
`, sess.AccountDID.String(), sess.SessionID, string(payload))
	return err
}

func (s *SQLiteStore) DeleteSession(ctx context.Context, did syntax.DID, sessionID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM oauth_sessions WHERE account_did = ? AND session_id = ?`,
		did.String(), sessionID,
	)
	return err
}

func (s *SQLiteStore) GetAuthRequestInfo(ctx context.Context, state string) (*indigooauth.AuthRequestData, error) {
	var payload string
	err := s.db.QueryRowContext(ctx,
		`SELECT payload FROM oauth_auth_requests WHERE state = ?`, state,
	).Scan(&payload)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("request info not found: %s", state)
	}
	if err != nil {
		return nil, err
	}
	var info indigooauth.AuthRequestData
	if err := json.Unmarshal([]byte(payload), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *SQLiteStore) SaveAuthRequestInfo(ctx context.Context, info indigooauth.AuthRequestData) error {
	payload, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO oauth_auth_requests (state, payload) VALUES (?, ?)
`, info.State, string(payload))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
			return fmt.Errorf("auth request already saved for state %s", info.State)
		}
		return err
	}
	return nil
}

func (s *SQLiteStore) DeleteAuthRequestInfo(ctx context.Context, state string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM oauth_auth_requests WHERE state = ?`, state)
	return err
}

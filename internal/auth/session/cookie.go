package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const CookieName = "twisky_session"

var (
	ErrMissingCookie = errors.New("session cookie missing")
	ErrInvalidCookie = errors.New("session cookie invalid")
)

// Account is one logged-in ATProto account bound to an indigo OAuth session id.
type Account struct {
	DID       string `json:"did"`
	SessionID string `json:"session_id"`
	Handle    string `json:"handle,omitempty"`
}

// State is the signed browser session payload (multi-account).
type State struct {
	ActiveDID string    `json:"active_did"`
	Accounts  []Account `json:"accounts"`
}

func (s State) ActiveAccount() (Account, bool) {
	for _, account := range s.Accounts {
		if account.DID == s.ActiveDID {
			return account, true
		}
	}
	return Account{}, false
}

// AddAccount upserts by DID and makes it active.
func (s State) AddAccount(account Account) State {
	accounts := make([]Account, 0, len(s.Accounts)+1)
	replaced := false
	for _, existing := range s.Accounts {
		if existing.DID == account.DID {
			accounts = append(accounts, account)
			replaced = true
			continue
		}
		accounts = append(accounts, existing)
	}
	if !replaced {
		accounts = append(accounts, account)
	}
	s.Accounts = accounts
	s.ActiveDID = account.DID
	return s
}

// RemoveAccount drops a DID; if it was active, promotes the first remaining account.
func (s State) RemoveAccount(did string) State {
	accounts := make([]Account, 0, len(s.Accounts))
	for _, account := range s.Accounts {
		if account.DID == did {
			continue
		}
		accounts = append(accounts, account)
	}
	s.Accounts = accounts
	if s.ActiveDID == did {
		s.ActiveDID = ""
		if len(accounts) > 0 {
			s.ActiveDID = accounts[0].DID
		}
	}
	return s
}

// Jar signs and verifies session cookies with HMAC-SHA256.
type Jar struct {
	secret []byte
	secure bool
	maxAge time.Duration
}

func NewJar(secret []byte) *Jar {
	return &Jar{
		secret: secret,
		secure: true,
		maxAge: 30 * 24 * time.Hour,
	}
}

// SetSecure controls the Secure cookie flag (false for plain HTTP localhost).
func (j *Jar) SetSecure(secure bool) {
	j.secure = secure
}

func (j *Jar) Load(r *http.Request) (State, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return State{}, ErrMissingCookie
	}
	return j.decode(cookie.Value)
}

func (j *Jar) Save(w http.ResponseWriter, state State) error {
	value, err := j.encode(state)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   j.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(j.maxAge.Seconds()),
	})
	return nil
}

func (j *Jar) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   j.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (j *Jar) encode(state State) (string, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, j.secret)
	_, _ = mac.Write([]byte(payloadB64))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payloadB64 + "." + sigB64, nil
}

func (j *Jar) decode(value string) (State, error) {
	payloadB64, sigB64, ok := strings.Cut(value, ".")
	if !ok || payloadB64 == "" || sigB64 == "" {
		return State{}, ErrInvalidCookie
	}
	mac := hmac.New(sha256.New, j.secret)
	_, _ = mac.Write([]byte(payloadB64))
	want, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return State{}, fmt.Errorf("%w: signature encoding", ErrInvalidCookie)
	}
	if !hmac.Equal(mac.Sum(nil), want) {
		return State{}, ErrInvalidCookie
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return State{}, fmt.Errorf("%w: payload encoding", ErrInvalidCookie)
	}
	var state State
	if err := json.Unmarshal(payload, &state); err != nil {
		return State{}, fmt.Errorf("%w: payload json", ErrInvalidCookie)
	}
	return state, nil
}

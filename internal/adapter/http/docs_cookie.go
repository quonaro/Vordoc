package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const accessCookieName = "vordoc_access"

// unlockedScope is a single scope entry in the access cookie.
type unlockedScope struct {
	Exp int64 `json:"exp"`
}

// cookieValue holds all unlocked scopes per documentation.
type cookieValue struct {
	Scopes map[string]map[string]unlockedScope `json:"scopes"`
}

func (h *DocsHandler) parseCookie(r *http.Request) (cookieValue, bool) {
	cookie, err := r.Cookie(accessCookieName)
	if err != nil {
		return cookieValue{}, false
	}

	val, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return cookieValue{}, false
	}

	parts := splitSignedValue(string(val))
	if len(parts) != 2 {
		return cookieValue{}, false
	}

	sigBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return cookieValue{}, false
	}

	expectedSig := h.sign(parts[1])
	if !hmac.Equal(sigBytes, expectedSig) {
		return cookieValue{}, false
	}

	var cv cookieValue
	if err := json.Unmarshal([]byte(parts[1]), &cv); err != nil {
		return cookieValue{}, false
	}

	if cv.Scopes == nil {
		cv.Scopes = make(map[string]map[string]unlockedScope)
	}

	now := time.Now().Unix()
	for doc, scopes := range cv.Scopes {
		for scope, entry := range scopes {
			if entry.Exp < now {
				delete(scopes, scope)
			}
		}
		if len(scopes) == 0 {
			delete(cv.Scopes, doc)
		}
	}

	return cv, true
}

func (h *DocsHandler) hasValidCookie(r *http.Request, doc, scope string) bool {
	cv, ok := h.parseCookie(r)
	if !ok {
		return false
	}
	entry, ok := cv.Scopes[doc][scope]
	if !ok {
		return false
	}
	return entry.Exp >= time.Now().Unix()
}

func (h *DocsHandler) setAccessCookie(w http.ResponseWriter, r *http.Request, doc, scope string) {
	cv, _ := h.parseCookie(r)
	if cv.Scopes == nil {
		cv.Scopes = make(map[string]map[string]unlockedScope)
	}
	if cv.Scopes[doc] == nil {
		cv.Scopes[doc] = make(map[string]unlockedScope)
	}
	cv.Scopes[doc][scope] = unlockedScope{
		Exp: time.Now().Add(24 * time.Hour).Unix(),
	}

	data, _ := json.Marshal(cv)
	sig := base64.URLEncoding.EncodeToString(h.sign(string(data)))
	raw := fmt.Sprintf("%s.%s", sig, string(data))
	value := base64.URLEncoding.EncodeToString([]byte(raw))

	// #nosec G124 — Secure отключён, приложение может работать по HTTP; HttpOnly и SameSite установлены.
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
}

func (h *DocsHandler) sign(data string) []byte {
	mac := hmac.New(sha256.New, h.cookieSecret)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func splitSignedValue(val string) []string {
	for i := 0; i < len(val); i++ {
		if val[i] == '.' {
			return []string{val[:i], val[i+1:]}
		}
	}
	return nil
}

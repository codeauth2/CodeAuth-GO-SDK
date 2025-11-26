package CodeAuthSDK

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	// Global instance (like JS static class variables)
	Endpoint       string
	ProjectID      string
	UseCache       bool
	CacheDuration  time.Duration
	cacheSession   map[string]map[string]interface{}
	cacheTimestamp time.Time
	hasInitialized bool

	mutex sync.Mutex
	client = &http.Client{}
)

var ErrNotInitialized = errors.New("CodeAuth has not been initialized")
var ErrAlreadyInitialized = errors.New("CodeAuth has already been initialized")

// Initialize 
func Initialize(projectEndpoint, projectID string, useCache bool, cacheSeconds int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if hasInitialized {
		return ErrAlreadyInitialized
	}
	hasInitialized = true

	Endpoint = projectEndpoint
	ProjectID = projectID
	UseCache = useCache
	CacheDuration = time.Duration(cacheSeconds) * time.Second
	cacheTimestamp = time.Now()
	cacheSession = make(map[string]map[string]interface{})

	return nil
}

// --------------------------------------------
// Internal POST request
// --------------------------------------------
func request(path string, body map[string]interface{}) map[string]interface{} {
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://"+Endpoint+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &out); err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}

	if resp.StatusCode == 200 {
		out["error"] = "no_error"
	}

	return out
}

// --------------------------------------------
// API Methods
// --------------------------------------------

func SignInEmail(email string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	body := map[string]interface{}{
		"project_id": ProjectID,
		"email":      email,
	}

	return request("/signin/email", body), nil
}

func SignInEmailVerify(email, code string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	return request("/signin/emailverify", map[string]interface{}{
		"project_id": ProjectID,
		"email":      email,
		"code":       code,
	}), nil
}

func SignInSocial(socialType string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	return request("/signin/social", map[string]interface{}{
		"project_id":  ProjectID,
		"social_type": socialType,
	}), nil
}

func SignInSocialVerify(socialType, authorizationCode string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	return request("/signin/socialverify", map[string]interface{}{
		"project_id":        ProjectID,
		"social_type":       socialType,
		"authorization_code": authorizationCode,
	}), nil
}

func SessionInfo(sessionToken string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Cache hit
	if UseCache {
		if time.Since(cacheTimestamp) < CacheDuration {
			if v, ok := cacheSession[sessionToken]; ok {
				return v, nil
			}
		} else {
			cacheTimestamp = time.Now()
			cacheSession = make(map[string]map[string]interface{})
		}
	}

	resp := request("/session/info", map[string]interface{}{
		"project_id":    ProjectID,
		"session_token": sessionToken,
	})

	// Save cache
	if UseCache && resp["error"] == "no_error" {
		cacheSession[sessionToken] = resp
	}

	return resp, nil
}

func SessionRefresh(sessionToken string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	mutex.Lock()
	defer mutex.Unlock()

	resp := request("/session/refresh", map[string]interface{}{
		"project_id":    ProjectID,
		"session_token": sessionToken,
	})

	if UseCache && resp["error"] != "no_error" {
		if time.Since(cacheTimestamp) >= CacheDuration {
			cacheTimestamp = time.Now()
			cacheSession = make(map[string]map[string]interface{})
		} else {
			// delete old token
			delete(cacheSession, sessionToken)

			// insert new one
			if newToken, ok := resp["session_token"].(string); ok {
				cacheSession[newToken] = map[string]interface{}{
					"email":        resp["email"],
					"expiration":   resp["expiration"],
					"refresh_left": resp["refresh_left"],
				}
			}
		}
	}

	return resp, nil
}

func SessionInvalidate(sessionToken, invalidateType string) (map[string]interface{}, error) {
	if !hasInitialized {
		return nil, ErrNotInitialized
	}

	mutex.Lock()
	defer mutex.Unlock()

	resp := request("/session/invalidate", map[string]interface{}{
		"project_id":     ProjectID,
		"session_token":  sessionToken,
		"invalidate_type": invalidateType,
	})

	if UseCache && resp["error"] != "no_error" {
		if time.Since(cacheTimestamp) >= CacheDuration {
			cacheTimestamp = time.Now()
			cacheSession = make(map[string]map[string]interface{})
		} else {
			delete(cacheSession, sessionToken)
		}
	}

	return resp, nil
}
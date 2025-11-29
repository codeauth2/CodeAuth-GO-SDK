package CodeAuth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var (
	endpoint        string
	projectID       string
	useCache        bool
	cacheDuration   time.Duration
	cacheSession    map[string]map[string]interface{}
	cacheExpiration time.Time
	hasInitialized  bool
	mutex           sync.Mutex
	client = &http.Client{}
)

var ErrNotInitialized = errors.New("CodeAuth has not been initialized")
var ErrAlreadyInitialized = errors.New("CodeAuth has already been initialized")

// --------------------------------------------
// Initialize the CodeAuth SDK
// project_endpoint - The endpoint of your project. This can be found inside your project settings.
// project_id - Your project ID. This can be found inside your project settings.
// use_cache - Whether to use cache or not. Using cache can help speed up response time and mitigate some rate limits. This will automatically cache new session token (from '/signin/emailverify', 'signin/socialverify', 'session/info', 'session/refresh') and automatically delete cache when it is invalidated (from 'session/refresh', 'session/invalidate').
// cache_duration - How long the cache should last. At least 15 seconds required to effectively mitigate most rate limits. Check docs for more info.
// --------------------------------------------
func Initialize(project_endpoint string, project_id string, use_cache bool, cache_duration int) {
	mutex.Lock()
	defer mutex.Unlock()

	if hasInitialized {
		panic("CodeAuth has already been initialized.")
	}

	hasInitialized = true
	endpoint = project_endpoint
	projectID = project_id
	useCache = use_cache
	cacheDuration = time.Duration(cache_duration) * time.Second
	cacheSession = make(map[string]map[string]interface{})
	cacheExpiration = time.Now().Add(cacheDuration)
}

// Makes sure that the CodeAuth SDK has been initialized
func ensureInitialized() {
	if !hasInitialized {
		panic("CodeAuth has not been initialized.")
	}
}

// Makes sure cache hasn't expired, if it did, delete the whole map
func ensureCache() {
	if !useCache {
		return
	}
	if time.Now().After(cacheExpiration) {
		cacheExpiration = time.Now().Add(cacheDuration)
		cacheSession = make(map[string]map[string]interface{})
	}
}

// Create api request and call server
func callApiRequest(path string, body map[string]interface{}) map[string]interface{} {
	defer func() {
		recover() 
	}()

	fullURL := "https://" + endpoint + path

	dataBytes, err := json.Marshal(body)
	if err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(dataBytes))
	if err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}

	var jsonResp map[string]interface{}
	if err := json.Unmarshal(respData, &jsonResp); err != nil {
		return map[string]interface{}{"error": "connection_error"}
	}

	if resp.StatusCode == 200 {
		jsonResp["error"] = "no_error"
	}

	return jsonResp
}

// --------------------------------------------
// @summary Begins the sign in or register flow by sending the user a one time code via email.
// @param {string} email - The email of the user you are trying to sign in/up. Email must be between 1 and 64 characters long. The email must also only contain letter, number, dot (not first, last, or consecutive), underscore(not first or last) and/or hyphen(not first or last).
// @returns A success response will return error = 'no_error'
// --------------------------------------------
func SignInEmail(email string) map[string]interface{}{
	// make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()
	mutex.Unlock()

	// return signin email 
	return callApiRequest("/signin/email", map[string]interface{}{
		"project_id": projectID,
		"email":      email,
	})
}

// --------------------------------------------
// @summary Checks if the one time code matches in order to create a session token.
// @param {string} email - The email of the user you are trying to sign in/up. Email must be between 1 and 64 characters long. The email must also only contain letter, number, dot (not first, last, or consecutive), underscore(not first or last) and/or hyphen(not first or last).
// @param {string} code - The one time code that was sent to the email.
// @returns {object} { session_token, email, expiration, refresh_left } 
// --------------------------------------------
func SignInEmailVerify(email, code string) map[string]interface{} {
    // make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()
	mutex.Unlock()

	// call server and get response 
	result := callApiRequest("/signin/emailverify", map[string]interface{}{
		"project_id": projectID,
		"email":      email,
		"code":       code,
	})

	// save to cache if enabled
	if useCache && result["error"] == "no_error" {
		mutex.Lock()
		cacheSession[result["session_token"].(string)] = result
		mutex.Unlock()
	}

	// return signin email verify
	return result
}

// --------------------------------------------
// @summary Begins the sign in or register flow by allowing users to sign in through a social OAuth2 link.
// @param {string} social_type - The type of social OAuth2 url you are trying to create. Possible social types: "google", "microsoft", "apple"
// @returns {object} { signin_url }
// --------------------------------------------
func SignInSocial(socialType string) map[string]interface{} {
    // make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()
	mutex.Unlock()

	// return signin social 
	return callApiRequest("/signin/social", map[string]interface{}{
		"project_id":  projectID,
		"social_type": socialType,
	})
}

// --------------------------------------------
// @summary This is the next step after the user signs in with their social account. This request checks the authorization code given by the social media company in order to create a session token.
// @param {string} social_type - The type of social OAuth2 url you are trying to verify. Possible social types: "google", "microsoft", "apple"
// @param {string} authorization_code - The authorization code given by the social. Check the docs for more info.
// @returns {object} { session_token, email, expiration, refresh_left }
// --------------------------------------------
func SignInSocialVerify(socialType, authorizationCode string) map[string]interface{} {
    // make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()
	mutex.Unlock()

	// call server and get response 
	result := callApiRequest("/signin/socialverify", map[string]interface{}{
		"project_id":         projectID,
		"social_type":        socialType,
		"authorization_code": authorizationCode,
	})

	// save to cache if enabled
	if useCache && result["error"] == "no_error" {
		mutex.Lock()
		cacheSession[result["session_token"].(string)] = result
		mutex.Unlock()
	}

	// return signin social verify
	return result
}

// --------------------------------------------
// @summary Gets the information associated with a session token.
// @param {string} session_token - The session token you are trying to get information on.
// @returns {object} { email, expiration, refresh_left }
// --------------------------------------------
func SessionInfo(sessionToken string) map[string]interface{} {
    // make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()

	// return the cached info if it is enabled, not expired and exist
	if useCache && time.Now().Before(cacheExpiration) {
		if cached, ok := cacheSession[sessionToken]; ok {
			mutex.Unlock()
			return cached
		}
	}
	mutex.Unlock()

	// call server and get response 
	result := callApiRequest("/session/info", map[string]interface{}{
		"project_id":    projectID,
		"session_token": sessionToken,
	})

	// save to cache if enabled
	if useCache && result["error"] == "no_error" {
		mutex.Lock()
		cacheSession[sessionToken] = result
		mutex.Unlock()
	}

	// return session info
	return result
}

// --------------------------------------------
// @summary Create a new session token using existing session token.
// @param {string} session_token - The session token you are trying to use to create a new token.
// @returns {object} { session_token:<string>, email:<string>, expiration:<int>, refresh_left:<int> }
// --------------------------------------------
func SessionRefresh(sessionToken string) map[string]interface{} {
    // make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()
	mutex.Unlock()

	// call server and get response
	result := callApiRequest("/session/refresh", map[string]interface{}{
		"project_id":    projectID,
		"session_token": sessionToken,
	})

	// if cache is enabled, delete old session token cache and set the new one
	if useCache && result["error"] == "no_error" {
		newToken := result["session_token"].(string)

		mutex.Lock()
		delete(cacheSession, sessionToken)
		cacheSession[newToken] = result
		mutex.Unlock()
	}

	// return
	return result
}

// --------------------------------------------
// @summary Invalidate a session token. By doing so, the session token can no longer be used for any api call.
// @param {string} session_token - The session token you are trying to use to invalidate.
// @param {string} invalidate_type - How to use the session token to invalidate. Possible invalidate types: 'only_this', 'all', 'all_but_this'
// @returns {object} {}
// --------------------------------------------
func SessionInvalidate(sessionToken, invalidateType string) map[string]interface{}{
    // make sure CodeAuth SDK has been initialized
	ensureInitialized()

	// make sure cache if valid
	mutex.Lock()
	ensureCache()
	mutex.Unlock()

	// call server and get response 
	result := callApiRequest("/session/invalidate", map[string]interface{}{
		"project_id":     projectID,
		"session_token":  sessionToken,
		"invalidate_type": invalidateType,
	})

	// if cache is enabled, and there is no problem with the request, delete the session token cache
	if useCache && result["error"] == "no_error" {
		mutex.Lock()
		delete(cacheSession, sessionToken)
		mutex.Unlock()
	}

	// return
	return result
}

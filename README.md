# CodeAuth GO SDK
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/codeauth2/CodeAuth-GO-SDK)

Offical CodeAuth SDK. For more info, check the docs on our [official website](https://docs.codeauth.com).

## Install
`go get github.com/codeauth2/CodeAuth-GO-SDK`

## Basic Usage

### Initialize CodeAuth SDK
```go
import "github.com/codeauth2/CodeAuth-GO-SDK"
CodeAuth.Initialize("<your project API endpoint>", "<your project ID>")
```

### Signin / Email
Begins the sign in or register flow by sending the user a one time code via email.
```go
result := CodeAuth.SignInEmail("<user email>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_email": 
	case "code_request_interval_reached": 
	case "code_hourly_limit_reached": 
	case "email_provider_error": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server
}
```

### Signin / Email Verify
Checks if the one time code matches in order to create a session token.
```go
result := CodeAuth.SignInEmailVerify("<user email>", "<one time code>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_email": 
	case "bad_code": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server
}
print(result["session_token"])
print(result["email"])
print(result["expiration"])
print(result["refresh_left"])
```

### Signin / Social
Begins the sign in or register flow by allowing users to sign in through a social OAuth2 link.
```go
result := CodeAuth.SignInSocial("<social_type>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_social_type": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server
}
print(result["signin_url"])
```

### Signin / Social Verify
This is the next step after the user signs in with their social account. This request checks the authorization code given by the social media company in order to create a session token.
```go
result := CodeAuth.SignInSocialVerify("<social type>", "<authorization code>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_social_type": 
	case "bad_authorization_code": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server
}
print(result["session_token"])
print(result["email"])
print(result["expiration"])
print(result["refresh_left"])
```

### Session / Info
Gets the information associated with a session token.
```go
result := CodeAuth.SessionInfo("<session_token>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_session_token": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server
}
print(result["email"])
print(result["expiration"])
print(result["refresh_left"])
```

### Session / Refresh
Create a new session token using existing session token.
```go
result := CodeAuth.SessionRefresh("<session_token>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_session_token": 
	case "out_of_refresh": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server
}
print(result["session_token"])
print(result["email"])
print(result["expiration"])
print(result["refresh_left"])
```

### Session / Invalidate
Invalidate a session token. By doing so, the session token can no longer be used for any api call.
```go
result := CodeAuth.SessionInvalidate("<session_token>", "<invalidate_type>")
switch result["error"]{
	case "bad_json": 
	case "project_not_found": 
	case "bad_ip_address": 
	case "rate_limit_reached": 
	case "bad_session_token": 
	case "bad_invalidate_type": 
	case "internal_error": 
	case "connection_error": //sdk failed to connect to api server 
}
```

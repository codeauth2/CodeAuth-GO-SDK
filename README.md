# CodeAuth GO SDK

Offical CodeAuth SDK. For more info, check the docs on the [official website](https://docs.codeauth.com).

## Install
`go get github.com/codeauth2/CodeAuth-GO-SDK`

## Sample Usage
```go
import "github.com/codeauth2/CodeAuth-GO-SDK"
CodeAuthSDK.Initialize("<your project API endpoint>", "<your project ID>", true, 5)
result, _ := CodeAuth.SignInEmail("<user email>")
print(result)
```



# Configuration

Configuration file is used by UnDistro just in install and move operations.

## Reference

```go
type Config struct {
	Credentials   Credentials `mapstructure:"credentials" json:"credentials,omitempty"`
	CoreProviders []Provider  `mapstructure:"coreProviders" json:"coreProviders,omitempty"`
	Providers     []Provider  `mapstructure:"providers" json:"providers,omitempty"`
}
type Credentials struct {
	Username string `mapstructure:"username" json:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty"`
}

type Provider struct {
	Name          string            `mapstructure:"name" json:"name,omitempty"`
	Configuration map[string]string `mapstructure:"configuration" json:"configuration,omitempty"`
}
```

### Config

|Name       |Type       |Description|
|-----------|-----------|-----------|
|credentials|Credentials|The registry credentials to use private images|
|coreProviders|[]Provider|Core providers can be undistro, cert-manager, cluster-api|
|providers|[]Provider| providers can configure any supported infrastruture provider|

### Credentials

|Name       |Type       |Description|
|-----------|-----------|-----------|
|username|string|The registry username|
|password|string|The registry password|

### Provider

|Name       |Type       |Description|
|-----------|-----------|-----------|
|name|string|Provider name|
|configuration|map[string]string|Change according provider name. See provider docs|
package config

// Config provides configuration.
type Config struct {
	// SessionName is the cookie name for the session
	SessionName string

	// UserIDKey is the user_id session key
	UserIDKey string

	// ProtectedAreaDirExternal is the URL path of the protected area visible to the user.
	ProtectedAreaDirExternal string
	// ProtectedAreaDirInternal is the URL path of the protected area not visible to the user.
	ProtectedAreaDirInternal string
	// ProtectedAreaHome is the URL of the protected area's homepage.
	ProtectedAreaHome string
}

// NewConfig returns a new configuration with default values.
func NewConfig() *Config {
	return &Config{
		SessionName:              "auth-static",
		UserIDKey:                "user_id",
		ProtectedAreaDirExternal: "/private/",
		ProtectedAreaDirInternal: "/internal/",
		ProtectedAreaHome:        "main.html",
	}
}

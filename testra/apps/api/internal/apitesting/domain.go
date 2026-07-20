package apitesting

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HTTPMethod represents an allowed HTTP method for an API request.
type HTTPMethod string

const (
	MethodGET     HTTPMethod = "GET"
	MethodPOST    HTTPMethod = "POST"
	MethodPUT     HTTPMethod = "PUT"
	MethodPATCH   HTTPMethod = "PATCH"
	MethodDELETE  HTTPMethod = "DELETE"
	MethodHEAD    HTTPMethod = "HEAD"
	MethodOPTIONS HTTPMethod = "OPTIONS"
)

// IsValidHTTPMethod returns true for supported HTTP methods.
func IsValidHTTPMethod(m HTTPMethod) bool {
	switch m {
	case MethodGET, MethodPOST, MethodPUT, MethodPATCH, MethodDELETE, MethodHEAD, MethodOPTIONS:
		return true
	}
	return false
}

// AuthType represents the authentication scheme for an API request.
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeAPIKey AuthType = "api_key"
)

// IsValidAuthType returns true for supported auth types.
func IsValidAuthType(t AuthType) bool {
	switch t {
	case AuthTypeNone, AuthTypeBearer, AuthTypeBasic, AuthTypeAPIKey:
		return true
	}
	return false
}

// BodyType represents how the request body should be interpreted.
type BodyType string

const (
	BodyTypeNone       BodyType = "none"
	BodyTypeJSON       BodyType = "json"
	BodyTypeRaw        BodyType = "raw"
	BodyTypeForm       BodyType = "form"
	BodyTypeURLEncoded BodyType = "urlencoded"
)

// IsValidBodyType returns true for supported body types.
func IsValidBodyType(t BodyType) bool {
	switch t {
	case BodyTypeNone, BodyTypeJSON, BodyTypeRaw, BodyTypeForm, BodyTypeURLEncoded:
		return true
	}
	return false
}

// KeyValuePair is a reusable key/value toggle used for headers, query
// parameters, and variables.
type KeyValuePair struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// AuthConfig holds the values required by the selected auth_type.
type AuthConfig struct {
	BearerToken string `json:"bearer_token,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	APIKey      string `json:"api_key,omitempty"`
	APIValue    string `json:"api_value,omitempty"`
	APILocation string `json:"api_location,omitempty"` // "header" or "query"
}

// MarshalJSON is a thin wrapper so the config can be stored as JSONB.
func (a AuthConfig) MarshalJSON() ([]byte, error) {
	type alias AuthConfig
	return json.Marshal(alias(a))
}

// Collection groups related API requests within a workspace.
type Collection struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Description string
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Folder provides hierarchical organization of requests within a collection.
type Folder struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	CollectionID uuid.UUID
	ParentID     *uuid.UUID
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Environment is a set of variables that can be applied to a request.
type Environment struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Variables   []KeyValuePair
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Request is the core API test definition.
type Request struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	CollectionID  uuid.UUID
	FolderID      *uuid.UUID
	EnvironmentID *uuid.UUID
	Name          string
	Method        HTTPMethod
	URL           string
	Headers       []KeyValuePair
	QueryParams   []KeyValuePair
	AuthType      AuthType
	AuthConfig    AuthConfig
	BodyType      BodyType
	BodyContent   string
	Variables     []KeyValuePair
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RequestHistory records a single execution of an API request.
type RequestHistory struct {
	ID                 uuid.UUID
	WorkspaceID        uuid.UUID
	RequestID          *uuid.UUID
	EnvironmentID      *uuid.UUID
	Name               string
	Method             string
	URL                string
	RequestHeaders     []KeyValuePair
	RequestBody        string
	ResponseStatus     int
	ResponseStatusText string
	ResponseHeaders    []KeyValuePair
	ResponseBody       string
	ResponseTimeMs     int
	Error              string
	CreatedBy          uuid.UUID
	CreatedAt          time.Time
}

// ExecutionResult is the service-layer output of an API request execution.
type ExecutionResult struct {
	Status         int
	StatusText     string
	Headers        map[string][]string
	Body           []byte
	ResponseTimeMs int
	Error          string
}

// IsValidName is reused from shared validation; copied to avoid package import
// cycles and keep the package self-contained.
func IsValidName(name string) bool {
	return len(strings.TrimSpace(name)) >= 2 && len(name) <= 255
}

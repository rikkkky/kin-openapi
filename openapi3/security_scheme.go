package openapi3

import (
	"context"
	"errors"
	"fmt"

	"github.com/getkin/kin-openapi/jsoninfo"
)

type SecurityScheme struct {
	ExtensionProps

	Type         string      `json:"type,omitempty"`
	Description  string      `json:"description,omitempty"`
	Name         string      `json:"name,omitempty"`
	In           string      `json:"in,omitempty"`
	Scheme       string      `json:"scheme,omitempty"`
	BearerFormat string      `json:"bearerFormat,omitempty"`
	Flow         *OAuthFlows `json:"flow,omitempty"`
}

func NewSecurityScheme() *SecurityScheme {
	return &SecurityScheme{}
}

func NewCSRFSecurityScheme() *SecurityScheme {
	return &SecurityScheme{
		Type: "apiKey",
		In:   "header",
		Name: "X-XSRF-TOKEN",
	}
}

func NewJWTSecurityScheme() *SecurityScheme {
	return &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
}

func (ss *SecurityScheme) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(ss)
}

func (ss *SecurityScheme) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, ss)
}

func (ss *SecurityScheme) WithType(value string) *SecurityScheme {
	ss.Type = value
	return ss
}

func (ss *SecurityScheme) WithDescription(value string) *SecurityScheme {
	ss.Description = value
	return ss
}

func (ss *SecurityScheme) WithName(value string) *SecurityScheme {
	ss.Name = value
	return ss
}

func (ss *SecurityScheme) WithIn(value string) *SecurityScheme {
	ss.In = value
	return ss
}

func (ss *SecurityScheme) WithScheme(value string) *SecurityScheme {
	ss.Scheme = value
	return ss
}

func (ss *SecurityScheme) WithBearerFormat(value string) *SecurityScheme {
	ss.BearerFormat = value
	return ss
}

func (ss *SecurityScheme) Validate(c context.Context) error {
	hasIn := false
	hasBearerFormat := false
	hasFlow := false
	switch ss.Type {
	case "apiKey":
		hasIn = true
		hasBearerFormat = true
	case "http":
		scheme := ss.Scheme
		switch scheme {
		case "bearer":
			hasBearerFormat = true
		case "basic":
		default:
			return fmt.Errorf("Security scheme of type 'http' has invalid 'scheme' value '%s'", scheme)
		}
	case "oauth2":
		hasFlow = true
	case "openIdConnect":
		return fmt.Errorf("Support for security schemes with type '%v' has not been implemented", ss.Type)
	default:
		return fmt.Errorf("Security scheme 'type' can't be '%v'", ss.Type)
	}

	// Validate "in" and "name"
	if hasIn {
		switch ss.In {
		case "query", "header":
		default:
			return fmt.Errorf("Security scheme of type 'apiKey' should have 'in'. It can be 'query' or 'header', not '%s'", ss.In)
		}
		if ss.Name == "" {
			return errors.New("Security scheme of type 'apiKey' should have 'name'")
		}
	} else if len(ss.In) > 0 {
		return fmt.Errorf("Security scheme of type '%s' can't have 'in'", ss.Type)
	} else if len(ss.Name) > 0 {
		return errors.New("Security scheme of type 'apiKey' can't have 'name'")
	}

	// Validate "format"
	if hasBearerFormat {
		switch ss.BearerFormat {
		case "", "JWT":
		default:
			return fmt.Errorf("Security scheme has unsupported 'bearerFormat' value '%s'", ss.BearerFormat)
		}
	} else if len(ss.BearerFormat) > 0 {
		return errors.New("Security scheme of type 'apiKey' can't have 'bearerFormat'")
	}

	// Validate "flow"
	if hasFlow {
		flow := ss.Flow
		if flow == nil {
			return fmt.Errorf("Security scheme of type '%v' should have 'flow'", ss.Type)
		}
		if err := flow.Validate(c); err != nil {
			return fmt.Errorf("Security scheme 'flow' is invalid: %v", err)
		}
	} else if ss.Flow != nil {
		return fmt.Errorf("Security scheme of type '%s' can't have 'flow'", ss.Type)
	}
	return nil
}

type OAuthFlows struct {
	ExtensionProps
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

func (flows *OAuthFlows) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(flows)
}

func (flows *OAuthFlows) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, flows)
}

func (flows *OAuthFlows) Validate(c context.Context) error {
	if v := flows.Implicit; v != nil {
		return v.Validate(c)
	}
	if v := flows.Password; v != nil {
		return v.Validate(c)
	}
	if v := flows.ClientCredentials; v != nil {
		return v.Validate(c)
	}
	if v := flows.AuthorizationCode; v != nil {
		return v.Validate(c)
	}
	return errors.New("No OAuth flow is defined")
}

type OAuthFlow struct {
	ExtensionProps
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

func (flow *OAuthFlow) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(flow)
}

func (flow *OAuthFlow) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, flow)
}

func (flow *OAuthFlow) Validate(c context.Context) error {
	if v := flow.AuthorizationURL; v == "" {
		return errors.New("An OAuth flow is missing 'authorizationUrl'")
	}
	if v := flow.TokenURL; v == "" {
		return errors.New("An OAuth flow is missing 'tokenUrl'")
	}
	if v := flow.Scopes; len(v) == 0 {
		return errors.New("An OAuth flow is missing 'scopes'")
	}
	return nil
}

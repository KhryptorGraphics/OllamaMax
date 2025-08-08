package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &SwaggerSpec{
	Version:     "1.0.0",
	Host:        "localhost:8080",
	BasePath:    "/api/v1",
	Title:       "OllamaMax Distributed API",
	Description: "Enterprise-grade distributed AI inference platform with SSO and multi-tenancy",
	Contact: &Contact{
		Name:  "OllamaMax Team",
		Email: "support@ollamamax.com",
		URL:   "https://ollamamax.com",
	},
	License: &License{
		Name: "MIT",
		URL:  "https://opensource.org/licenses/MIT",
	},
}

// SwaggerSpec represents the OpenAPI specification
type SwaggerSpec struct {
	OpenAPI     string                `json:"openapi"`
	Info        *Info                 `json:"info"`
	Host        string                `json:"host,omitempty"`
	BasePath    string                `json:"basePath,omitempty"`
	Schemes     []string              `json:"schemes,omitempty"`
	Consumes    []string              `json:"consumes,omitempty"`
	Produces    []string              `json:"produces,omitempty"`
	Paths       map[string]*PathItem  `json:"paths"`
	Components  *Components           `json:"components,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
	Tags        []*Tag                `json:"tags,omitempty"`
	Version     string                `json:"-"`
	Title       string                `json:"-"`
	Description string                `json:"-"`
	Contact     *Contact              `json:"-"`
	License     *License              `json:"-"`
}

// Info represents API information
type Info struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Contact     *Contact `json:"contact,omitempty"`
	License     *License `json:"license,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// License represents license information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// PathItem represents a path item in the OpenAPI spec
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

// Operation represents an operation in the OpenAPI spec
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []*Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter represents a parameter in the OpenAPI spec
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody represents a request body in the OpenAPI spec
type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]*MediaType `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

// Response represents a response in the OpenAPI spec
type Response struct {
	Description string                `json:"description"`
	Content     map[string]*MediaType `json:"content,omitempty"`
	Headers     map[string]*Header    `json:"headers,omitempty"`
}

// MediaType represents a media type in the OpenAPI spec
type MediaType struct {
	Schema  *Schema     `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}

// Header represents a header in the OpenAPI spec
type Header struct {
	Description string      `json:"description,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Schema represents a schema in the OpenAPI spec
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
}

// Components represents components in the OpenAPI spec
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme in the OpenAPI spec
type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// Tag represents a tag in the OpenAPI spec
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GenerateSwaggerSpec generates the complete OpenAPI specification
func GenerateSwaggerSpec() *SwaggerSpec {
	spec := &SwaggerSpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:       SwaggerInfo.Title,
			Description: SwaggerInfo.Description,
			Version:     SwaggerInfo.Version,
			Contact:     SwaggerInfo.Contact,
			License:     SwaggerInfo.License,
		},
		Host:     SwaggerInfo.Host,
		BasePath: SwaggerInfo.BasePath,
		Schemes:  []string{"http", "https"},
		Consumes: []string{"application/json"},
		Produces: []string{"application/json"},
		Paths:    generatePaths(),
		Components: &Components{
			Schemas:         generateSchemas(),
			SecuritySchemes: generateSecuritySchemes(),
		},
		Security: []map[string][]string{
			{"BearerAuth": {}},
			{"ApiKeyAuth": {}},
		},
		Tags: generateTags(),
	}

	return spec
}

// generatePaths generates API paths
func generatePaths() map[string]*PathItem {
	paths := make(map[string]*PathItem)

	// Authentication endpoints
	paths["/auth/login"] = &PathItem{
		Post: &Operation{
			Tags:        []string{"Authentication"},
			Summary:     "User login",
			Description: "Authenticate user with username and password",
			OperationID: "login",
			RequestBody: &RequestBody{
				Description: "Login credentials",
				Required:    true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/LoginRequest"},
					},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "Login successful",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/AuthResponse"},
						},
					},
				},
				"401": {Description: "Invalid credentials"},
				"500": {Description: "Internal server error"},
			},
		},
	}

	// SSO endpoints
	paths["/auth/oauth2/{provider}/authorize"] = &PathItem{
		Get: &Operation{
			Tags:        []string{"SSO"},
			Summary:     "OAuth2 authorization",
			Description: "Initiate OAuth2 authorization flow",
			OperationID: "oauth2Authorize",
			Parameters: []*Parameter{
				{
					Name:        "provider",
					In:          "path",
					Description: "OAuth2 provider ID",
					Required:    true,
					Schema:      &Schema{Type: "string"},
				},
				{
					Name:        "redirect_url",
					In:          "query",
					Description: "Redirect URL after authorization",
					Required:    false,
					Schema:      &Schema{Type: "string"},
				},
			},
			Responses: map[string]*Response{
				"302": {Description: "Redirect to OAuth2 provider"},
				"400": {Description: "Invalid provider"},
				"500": {Description: "Internal server error"},
			},
		},
	}

	// Tenant management endpoints
	paths["/admin/tenants"] = &PathItem{
		Get: &Operation{
			Tags:        []string{"Tenant Management"},
			Summary:     "List tenants",
			Description: "Get list of all tenants",
			OperationID: "listTenants",
			Parameters: []*Parameter{
				{
					Name:        "status",
					In:          "query",
					Description: "Filter by tenant status",
					Required:    false,
					Schema:      &Schema{Type: "string"},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "List of tenants",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/TenantList"},
						},
					},
				},
			},
		},
		Post: &Operation{
			Tags:        []string{"Tenant Management"},
			Summary:     "Create tenant",
			Description: "Create a new tenant",
			OperationID: "createTenant",
			RequestBody: &RequestBody{
				Description: "Tenant creation request",
				Required:    true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/CreateTenantRequest"},
					},
				},
			},
			Responses: map[string]*Response{
				"201": {
					Description: "Tenant created",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/Tenant"},
						},
					},
				},
			},
		},
	}

	// Model management endpoints
	paths["/models"] = &PathItem{
		Get: &Operation{
			Tags:        []string{"Model Management"},
			Summary:     "List models",
			Description: "Get list of available models",
			OperationID: "listModels",
			Responses: map[string]*Response{
				"200": {
					Description: "List of models",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/ModelList"},
						},
					},
				},
			},
		},
	}

	// Inference endpoints
	paths["/inference"] = &PathItem{
		Post: &Operation{
			Tags:        []string{"Inference"},
			Summary:     "Run inference",
			Description: "Execute model inference",
			OperationID: "runInference",
			RequestBody: &RequestBody{
				Description: "Inference request",
				Required:    true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/InferenceRequest"},
					},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "Inference result",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{Ref: "#/components/schemas/InferenceResponse"},
						},
					},
				},
			},
		},
	}

	return paths
}

// generateSchemas generates component schemas
func generateSchemas() map[string]*Schema {
	schemas := make(map[string]*Schema)

	schemas["LoginRequest"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"username": {Type: "string", Description: "Username"},
			"password": {Type: "string", Description: "Password"},
		},
		Required: []string{"username", "password"},
	}

	schemas["AuthResponse"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"token":      {Type: "string", Description: "JWT token"},
			"expires_at": {Type: "string", Format: "date-time", Description: "Token expiration"},
			"user":       {Ref: "#/components/schemas/User"},
		},
	}

	schemas["User"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"id":       {Type: "string", Description: "User ID"},
			"username": {Type: "string", Description: "Username"},
			"email":    {Type: "string", Description: "Email address"},
			"role":     {Type: "string", Description: "User role"},
		},
	}

	schemas["Tenant"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"id":     {Type: "string", Description: "Tenant ID"},
			"name":   {Type: "string", Description: "Tenant name"},
			"domain": {Type: "string", Description: "Tenant domain"},
			"status": {Type: "string", Description: "Tenant status"},
		},
	}

	schemas["CreateTenantRequest"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name":        {Type: "string", Description: "Tenant name"},
			"domain":      {Type: "string", Description: "Tenant domain"},
			"admin_email": {Type: "string", Description: "Admin email"},
		},
		Required: []string{"name", "domain", "admin_email"},
	}

	schemas["TenantList"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"tenants": {
				Type:  "array",
				Items: &Schema{Ref: "#/components/schemas/Tenant"},
			},
			"total": {Type: "integer", Description: "Total number of tenants"},
		},
	}

	schemas["ModelList"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"models": {
				Type:  "array",
				Items: &Schema{Ref: "#/components/schemas/Model"},
			},
		},
	}

	schemas["Model"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"id":          {Type: "string", Description: "Model ID"},
			"name":        {Type: "string", Description: "Model name"},
			"description": {Type: "string", Description: "Model description"},
			"size":        {Type: "integer", Description: "Model size in bytes"},
		},
	}

	schemas["InferenceRequest"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"model":  {Type: "string", Description: "Model ID"},
			"prompt": {Type: "string", Description: "Input prompt"},
			"parameters": {
				Type:        "object",
				Description: "Inference parameters",
			},
		},
		Required: []string{"model", "prompt"},
	}

	schemas["InferenceResponse"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"response":   {Type: "string", Description: "Model response"},
			"model":      {Type: "string", Description: "Model used"},
			"created_at": {Type: "string", Format: "date-time", Description: "Response timestamp"},
		},
	}

	return schemas
}

// generateSecuritySchemes generates security schemes
func generateSecuritySchemes() map[string]*SecurityScheme {
	schemes := make(map[string]*SecurityScheme)

	schemes["BearerAuth"] = &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Bearer token authentication",
	}

	schemes["ApiKeyAuth"] = &SecurityScheme{
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "API key authentication",
	}

	return schemes
}

// generateTags generates API tags
func generateTags() []*Tag {
	return []*Tag{
		{Name: "Authentication", Description: "User authentication endpoints"},
		{Name: "SSO", Description: "Single Sign-On endpoints"},
		{Name: "Tenant Management", Description: "Multi-tenant management endpoints"},
		{Name: "Model Management", Description: "AI model management endpoints"},
		{Name: "Inference", Description: "AI inference endpoints"},
		{Name: "Monitoring", Description: "System monitoring endpoints"},
	}
}

// SwaggerHandler returns the Swagger JSON specification
func SwaggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := GenerateSwaggerSpec()
		c.JSON(http.StatusOK, spec)
	}
}

// SwaggerUIHandler returns the Swagger UI HTML
func SwaggerUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>OllamaMax API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/docs/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/json-to-terraform/parser/internal/handler" // register handlers
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/parser"
	"github.com/json-to-terraform/parser/internal/result"
)

// LambdaEvent is the invocation payload (e.g. from API Gateway).
type LambdaEvent struct {
	Body   string            `json:"body"`             // diagram JSON (raw or base64 if isBase64)
	IsBase64 bool            `json:"isBase64,omitempty"`
	EmitTfvars *bool         `json:"emitTfvars,omitempty"`
}

// LambdaResponse is returned to the client (API Gateway).
type LambdaResponse struct {
	StatusCode int               `json:"statusCode"`
	Success    bool              `json:"success"`
	Errors     []result.Error    `json:"errors,omitempty"`
	Warnings   []result.Warning  `json:"warnings,omitempty"`
	Files      map[string]string `json:"files,omitempty"` // filename -> content (base64)
}

// APIGatewayResponse is the shape expected by API Gateway proxy integration (body = JSON string).
type APIGatewayResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
}

func handler(ctx context.Context, event LambdaEvent) (APIGatewayResponse, error) {
	out := LambdaResponse{StatusCode: 200}

	body := event.Body
	if event.IsBase64 {
		dec, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			out.StatusCode = 400
			out.Success = false
			out.Errors = []result.Error{{Type: "invalid_input", Severity: "error", Message: "invalid base64 body: " + err.Error()}}
			return wrap(out), nil
		}
		body = string(dec)
	}

	var d diagram.Diagram
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		out.StatusCode = 400
		out.Success = false
		out.Errors = []result.Error{{Type: "invalid_json", Severity: "error", Message: "invalid diagram JSON: " + err.Error()}}
		return wrap(out), nil
	}

	opts := parser.DefaultOptions()
	if event.EmitTfvars != nil {
		opts.EmitTfvars = *event.EmitTfvars
	}
	p := parser.New(opts)
	res, err := p.Parse(&d)
	if err != nil {
		out.StatusCode = 500
		out.Success = false
		out.Errors = []result.Error{{Type: "parse_error", Severity: "error", Message: err.Error()}}
		return wrap(out), nil
	}

	out.Success = res.Success
	out.Errors = res.Errors
	out.Warnings = res.Warnings
	if res.Success && len(res.TerraformFiles) > 0 {
		out.Files = make(map[string]string)
		for name, content := range res.TerraformFiles {
			out.Files[name] = base64.StdEncoding.EncodeToString(content)
		}
	}
	if !res.Success {
		out.StatusCode = 422
	}
	return wrap(out), nil
}

func wrap(out LambdaResponse) APIGatewayResponse {
	bodyBytes, _ := json.Marshal(out)
	return APIGatewayResponse{
		StatusCode: out.StatusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(bodyBytes),
	}
}

func main() {
	lambda.Start(handler)
}

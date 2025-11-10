package builder

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type BuildRequest struct {
	FunctionName string
	Runtime      string // python, nodejs, go
	SourceCode   string // base64 encoded
	UserID       string
	Namespace    string
	Registry     string // Docker registry URL
}

type Builder struct {
	workDir  string
	registry string
}

func NewBuilder() *Builder {
	return &Builder{
		workDir:  "/tmp/builds",
		registry: os.Getenv("REGISTRY_URL"), // e.g., "registry.eventflow.local:5000"
	}
}

func (b *Builder) Build(ctx context.Context, req BuildRequest) (string, error) {
	buildID := uuid.New().String()[:8]
	buildDir := filepath.Join(b.workDir, buildID)

	// Create build directory
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir)

	// Decode source code
	sourceCode, err := base64.StdEncoding.DecodeString(req.SourceCode)
	if err != nil {
		return "", fmt.Errorf("failed to decode source code: %w", err)
	}

	// Generate Dockerfile based on runtime
	dockerfile, err := b.generateDockerfile(req.Runtime, string(sourceCode))
	if err != nil {
		return "", fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Write Dockerfile
	if err := os.WriteFile(filepath.Join(buildDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Write source code
	sourceFile := b.getSourceFileName(req.Runtime)
	if err := os.WriteFile(filepath.Join(buildDir, sourceFile), sourceCode, 0644); err != nil {
		return "", fmt.Errorf("failed to write source file: %w", err)
	}

	// Generate image tag
	imageTag := fmt.Sprintf("%s/%s-%s:%s", b.registry, req.Namespace, req.FunctionName, buildID)

	// Build image
	log.Printf("Building image: %s", imageTag)
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageTag, buildDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	// Push to registry
	log.Printf("Pushing image: %s", imageTag)
	cmd = exec.CommandContext(ctx, "docker", "push", imageTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to push image: %w", err)
	}

	log.Printf("✅ Successfully built and pushed: %s", imageTag)
	return imageTag, nil
}

func (b *Builder) generateDockerfile(runtime, sourceCode string) (string, error) {
	switch runtime {
	case "python":
		return b.generatePythonDockerfile(sourceCode), nil
	case "nodejs":
		return b.generateNodeJSDockerfile(sourceCode), nil
	case "go":
		return b.generateGoDockerfile(sourceCode), nil
	default:
		return "", fmt.Errorf("unsupported runtime: %s", runtime)
	}
}

func (b *Builder) generatePythonDockerfile(sourceCode string) string {
	// Detect if requirements.txt is needed
	hasRequirements := strings.Contains(sourceCode, "import") &&
		(strings.Contains(sourceCode, "requests") ||
			strings.Contains(sourceCode, "flask") ||
			strings.Contains(sourceCode, "fastapi"))

	dockerfile := `FROM python:3.11-slim

WORKDIR /app

`
	if hasRequirements {
		dockerfile += `# Install common dependencies
RUN pip install --no-cache-dir flask requests

`
	}

	dockerfile += `COPY handler.py .

ENV PORT=8080
EXPOSE 8080

CMD ["python", "handler.py"]
`
	return dockerfile
}

func (b *Builder) generateNodeJSDockerfile(sourceCode string) string {
	// Detect if package.json is needed
	hasPackages := strings.Contains(sourceCode, "require(") || strings.Contains(sourceCode, "import ")

	dockerfile := `FROM node:18-alpine

WORKDIR /app

`
	if hasPackages {
		dockerfile += `# Install common dependencies
RUN npm install express body-parser

`
	}

	dockerfile += `COPY handler.js .

ENV PORT=8080
EXPOSE 8080

CMD ["node", "handler.js"]
`
	return dockerfile
}

func (b *Builder) generateGoDockerfile(sourceCode string) string {
	return `FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY handler.go .

RUN go mod init function && \
    go mod tidy && \
    CGO_ENABLED=0 go build -o handler handler.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/handler .

ENV PORT=8080
EXPOSE 8080

CMD ["./handler"]
`
}

func (b *Builder) getSourceFileName(runtime string) string {
	switch runtime {
	case "python":
		return "handler.py"
	case "nodejs":
		return "handler.js"
	case "go":
		return "handler.go"
	default:
		return "handler"
	}
}

func main() {
	log.Println("Function Builder Service started")

	// This will be integrated with the API
	// For now, it's a standalone service that listens for build requests
	// via message queue or HTTP API

	builder := NewBuilder()

	// Example usage:
	ctx := context.Background()
	req := BuildRequest{
		FunctionName: "hello-python",
		Runtime:      "python",
		SourceCode: base64.StdEncoding.EncodeToString([]byte(`
from http.server import HTTPServer, BaseHTTPRequestHandler
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)
        
        # Process the request
        response = {"message": "Hello from Python!", "received": body.decode()}
        
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(response).encode())

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 8080), Handler)
    print('Function listening on port 8080')
    server.serve_forever()
`)),
		UserID:    "alice",
		Namespace: "tenant-alice",
		Registry:  "localhost:5000",
	}

	image, err := builder.Build(ctx, req)
	if err != nil {
		log.Fatalf("Build failed: %v", err)
	}

	log.Printf("Built image: %s", image)
}

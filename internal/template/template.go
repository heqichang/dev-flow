package template

import (
	"os"
	"path/filepath"
	"strings"
)

type Template struct {
	ProjectName string
	TemplateType string
	Author      string
	License     string
	Language    string
}

func NewTemplate(projectName, templateType, author, license, language string) *Template {
	return &Template{
		ProjectName:  projectName,
		TemplateType: templateType,
		Author:       author,
		License:      license,
		Language:     language,
	}
}

func (t *Template) Generate(targetDir string) error {
	files := t.getTemplateFiles()

	for path, content := range files {
		fullPath := filepath.Join(targetDir, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		processedContent := t.processTemplate(content)
		if err := os.WriteFile(fullPath, []byte(processedContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (t *Template) getTemplateFiles() map[string]string {
	files := map[string]string{
		"README.md": t.getReadme(),
		".gitignore": t.getGitignore(),
	}

	switch t.TemplateType {
	case "frontend":
		for k, v := range t.getFrontendTemplate() {
			files[k] = v
		}
	case "backend":
		for k, v := range t.getBackendTemplate() {
			files[k] = v
		}
	case "fullstack":
		for k, v := range t.getFullstackTemplate() {
			files[k] = v
		}
	case "cli":
		for k, v := range t.getCLITemplate() {
			files[k] = v
		}
	case "library":
		for k, v := range t.getLibraryTemplate() {
			files[k] = v
		}
	}

	return files
}

func (t *Template) processTemplate(content string) string {
	content = strings.ReplaceAll(content, "{{PROJECT_NAME}}", t.ProjectName)
	content = strings.ReplaceAll(content, "{{AUTHOR}}", t.Author)
	content = strings.ReplaceAll(content, "{{LICENSE}}", t.License)
	content = strings.ReplaceAll(content, "{{LANGUAGE}}", t.Language)
	return content
}

func (t *Template) getReadme() string {
	backtick := "`"
	return `# {{PROJECT_NAME}}

{{PROJECT_NAME}} 是一个 {{LANGUAGE}} 项目。

## 安装

根据你的环境和依赖管理工具安装项目依赖。

## 使用

使用 DevFlow CLI 管理项目:

` + backtick + backtick + backtick + `bash
# 查看可用脚本
devflow run

# 运行开发模式
devflow run dev

# 运行构建
devflow run build

# 运行测试
devflow run test
` + backtick + backtick + backtick + `

## 作者

{{AUTHOR}}

## 许可证

{{LICENSE}}
`
}

func (t *Template) getGitignore() string {
	return `# Node.js
node_modules/
dist/
build/
.env
.env.local
.env.*.local

# Go
vendor/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Python
__pycache__/
*.py[cod]
*.egg-info/
venv/
.env
.venv/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
logs/
*.log

# DevFlow
.devflow/
`
}

func (t *Template) getFrontendTemplate() map[string]string {
	return map[string]string{
		"src/index.js": `console.log("Hello from {{PROJECT_NAME}}!");
`,
		"src/App.js": `export default function App() {
  return <div>Hello, {{PROJECT_NAME}}!</div>;
}
`,
		"public/index.html": `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{PROJECT_NAME}}</title>
</head>
<body>
  <div id="root"></div>
</body>
</html>
`,
		"package.json": `{
  "name": "{{PROJECT_NAME}}",
  "version": "0.1.0",
  "description": "{{PROJECT_NAME}} 前端项目",
  "main": "src/index.js",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest"
  },
  "author": "{{AUTHOR}}",
  "license": "{{LICENSE}}",
  "devDependencies": {
    "vite": "^5.0.0",
    "vitest": "^1.0.0"
  }
}
`,
		"vite.config.js": `import { defineConfig } from 'vite';

export default defineConfig({
  server: {
    port: 3000,
    open: true
  },
  build: {
    outDir: 'dist',
    sourcemap: true
  }
});
`,
	}
}

func (t *Template) getBackendTemplate() map[string]string {
	return map[string]string{
		"main.go": `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from {{PROJECT_NAME}}!")
	})

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`,
		"go.mod": `module {{PROJECT_NAME}}

go 1.21
`,
		"internal/server/server.go": `package server

import "net/http"

type Server struct {
	mux *http.ServeMux
}

func New() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleRoot)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to {{PROJECT_NAME}} API"))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
`,
		"Makefile": `.PHONY: dev build test clean

dev:
	go run .

build:
	go build -o bin/{{PROJECT_NAME}} .

test:
	go test ./...

clean:
	rm -rf bin/
`,
	}
}

func (t *Template) getFullstackTemplate() map[string]string {
	return map[string]string{
		"server/main.go": `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	fmt.Println("API server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`,
		"client/package.json": `{
  "name": "{{PROJECT_NAME}}-client",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  },
  "dependencies": {},
  "devDependencies": {
    "vite": "^5.0.0"
  }
}
`,
		"client/src/index.js": `console.log("Hello from {{PROJECT_NAME}} client!");
`,
		"server/go.mod": `module {{PROJECT_NAME}}-server

go 1.21
`,
	}
}

func (t *Template) getCLITemplate() map[string]string {
	return map[string]string{
		"cmd/{{PROJECT_NAME}}/main.go": `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello from {{PROJECT_NAME}} CLI!")
	
	if len(os.Args) > 1 {
		fmt.Printf("Arguments: %v\n", os.Args[1:])
	}
}
`,
		"go.mod": `module {{PROJECT_NAME}}

go 1.21
`,
		"Makefile": `.PHONY: build install clean

build:
	go build -o bin/{{PROJECT_NAME}} ./cmd/{{PROJECT_NAME}}

install:
	go install ./cmd/{{PROJECT_NAME}}

clean:
	rm -rf bin/
`,
	}
}

func (t *Template) getLibraryTemplate() map[string]string {
	return map[string]string{
		"index.js": `/**
 * {{PROJECT_NAME}}
 * A utility library
 *
 * @author {{AUTHOR}}
 * @license {{LICENSE}}
 */

module.exports = {
  greet: function(name) {
    return 'Hello, ' + name + '!';
  },
  
  version: function() {
    return '0.1.0';
  }
};
`,
		"package.json": `{
  "name": "{{PROJECT_NAME}}",
  "version": "0.1.0",
  "description": "{{PROJECT_NAME}} library",
  "main": "index.js",
  "module": "index.js",
  "type": "commonjs",
  "scripts": {
    "test": "jest",
    "lint": "eslint ."
  },
  "author": "{{AUTHOR}}",
  "license": "{{LICENSE}}",
  "keywords": [
    "{{PROJECT_NAME}}",
    "library"
  ],
  "devDependencies": {
    "jest": "^29.0.0"
  }
}
`,
	}
}

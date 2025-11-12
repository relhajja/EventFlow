import { useState } from 'react';
import Editor from '@monaco-editor/react';

interface CodeEditorProps {
  onDeploy: (runtime: string, code: string) => void;
  initialRuntime?: string;
}

const CODE_TEMPLATES = {
  python: `from http.server import HTTPServer, BaseHTTPRequestHandler
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        response = {"message": "Hello from Python!"}
        self.wfile.write(json.dumps(response).encode())

if __name__ == '__main__':
    HTTPServer(('0.0.0.0', 8080), Handler).serve_forever()
`,
  nodejs: `const express = require('express');
const app = express();

app.use(express.json());

app.post('/', (req, res) => {
  res.json({ message: 'Hello from Node.js!' });
});

app.listen(8080, '0.0.0.0', () => {
  console.log('Server running on port 8080');
});
`,
  go: `package main

import (
    "encoding/json"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Hello from Go!",
    })
}

func main() {
    http.HandleFunc("/", handler)
    log.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
`
};

const RUNTIME_INFO = {
  python: {
    label: 'Python 3.11',
    language: 'python',
    icon: '🐍',
    description: 'Flask-based HTTP server'
  },
  nodejs: {
    label: 'Node.js 20',
    language: 'javascript',
    icon: '📦',
    description: 'Express.js web framework'
  },
  go: {
    label: 'Go 1.21',
    language: 'go',
    icon: '🚀',
    description: 'Native HTTP server'
  }
};

export default function CodeEditor({ onDeploy, initialRuntime = 'python' }: CodeEditorProps) {
  const [runtime, setRuntime] = useState<'python' | 'nodejs' | 'go'>(initialRuntime as any);
  const [code, setCode] = useState(CODE_TEMPLATES[runtime]);
  const [isDeploying, setIsDeploying] = useState(false);

  const handleRuntimeChange = (newRuntime: 'python' | 'nodejs' | 'go') => {
    setRuntime(newRuntime);
    setCode(CODE_TEMPLATES[newRuntime]);
  };

  const handleDeploy = async () => {
    setIsDeploying(true);
    try {
      await onDeploy(runtime, code);
    } finally {
      setIsDeploying(false);
    }
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6">
      {/* Runtime Selector */}
      <div className="flex items-center gap-2 mb-4">
        <label className="text-sm font-medium text-slate-300">Runtime:</label>
        <div className="flex gap-2">
          {(Object.keys(RUNTIME_INFO) as Array<'python' | 'nodejs' | 'go'>).map((rt) => {
            const info = RUNTIME_INFO[rt];
            return (
              <button
                key={rt}
                onClick={() => handleRuntimeChange(rt)}
                className={`px-4 py-2 rounded-lg border-2 transition-all ${
                  runtime === rt
                    ? 'border-primary-500 bg-primary-500/10 text-white'
                    : 'border-slate-600 bg-slate-900 text-slate-300 hover:border-slate-500'
                }`}
                title={info.description}
              >
                <span className="mr-2">{info.icon}</span>
                {info.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* Info Banner */}
      <div className="mb-4 p-3 bg-blue-900/20 border border-blue-500/30 rounded-lg">
        <div className="flex items-start gap-2">
          <span className="text-blue-400 text-lg">ℹ️</span>
          <div className="text-sm text-blue-300">
            <strong>Function Requirements:</strong> Your code must listen on <code className="bg-blue-900/50 px-1 rounded text-blue-200">0.0.0.0:8080</code>
            {' '}and respond to POST requests.
          </div>
        </div>
      </div>

      {/* Monaco Editor - MUCH BIGGER */}
      <div className="border-2 border-slate-600 rounded-lg overflow-hidden" style={{ height: '600px' }}>
        <Editor
          height="600px"
          language={RUNTIME_INFO[runtime].language}
          value={code}
          onChange={(value) => setCode(value || '')}
          theme="vs-dark"
          options={{
            minimap: { enabled: true },
            fontSize: 16,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            insertSpaces: true,
            wordWrap: 'on',
            folding: true,
            renderWhitespace: 'selection',
            lineHeight: 24,
            padding: { top: 16, bottom: 16 },
          }}
        />
      </div>

      {/* Action Buttons */}
      <div className="flex items-center justify-between mt-4">
        <div className="text-sm text-slate-400">
          {code.split('\n').length} lines • {code.length} characters
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setCode(CODE_TEMPLATES[runtime])}
            className="px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
          >
            Reset to Template
          </button>
          <button
            onClick={handleDeploy}
            disabled={isDeploying || !code.trim()}
            className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
          >
            {isDeploying ? (
              <span className="flex items-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Deploying...
              </span>
            ) : (
              'Deploy Function'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}

# multipilot

A simple orchestration layer based with Temporal Workflows to run multiple Copilot tasks at once.

## Prerequisites

- Temporal CLI installed, or a Temporal cluster running
- Go v1.24+
- GitHub Copilot CLI

## Installation

```bash
go install github.com/AstraBert/multipilot
```

## Usage

Start a Temporal server:

```bash
temporal server start-dev --db-filename multipilot.db
```

Start the temporal worker that will orchestrate all the tasks:

```bash
multipilot start-worker
```

Create a configuration file with all the tasks you want Copilot to perform, following this blueprint:

```json
{"tasks": 
  [
    {
      "log_file": "copilot-session-backend.log",
      "cwd": "/home/user/backend",
      "log_level": "info",
      "env": [
        "PATH=/usr/local/bin:/usr/bin:/bin",
        "NODE_ENV=production",
        "API_KEY=secret123"
      ],
      "prompt": "Analyze the codebase and suggest improvements for performance",
      "token": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "ai_model": "gpt-5.1",
      "system_prompt": "You are a helpful coding assistant specializing in Go and Python.",
      "exclude_tools": [
        "shell(rm)",
        "shell(git push)",
      ],
      "skills": [
        "code_review",
        "refactoring",
        "testing"
      ],
      "local_mcp_servers": {
        "filesystem": {
          "tools": ["read_file", "write_file", "list_directory"],
          "type": "stdio",
          "timeout": 30,
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/project"],
          "env": {
            "NODE_ENV": "production"
          },
          "cwd": "/home/user/project"
        },
        "database": {
          "tools": ["query", "execute"],
          "type": "local",
          "timeout": 60,
          "command": "/usr/local/bin/db-mcp-server",
          "args": ["--config", "/etc/db-config.json"],
          "env": {
            "DB_HOST": "localhost",
            "DB_PORT": "5432"
          }
        }
      },
      "remote_mcp_servers": {
        "weather_api": {
          "tools": ["get_weather", "get_forecast"],
          "type": "http",
          "timeout": 15,
          "url": "https://api.weather.example.com/mcp",
          "headers": {
            "Authorization": "Bearer token123",
            "Content-Type": "application/json"
          }
        },
        "analytics": {
          "tools": ["track_event", "get_metrics"],
          "type": "sse",
          "timeout": 45,
          "url": "https://analytics.example.com/mcp/stream",
          "headers": {
            "X-API-Key": "analytics_key_456"
          }
        }
      },
      "timeout_sec": 300
    },
    {
      "log_file": "copilot-session-frontend.log",
      "cwd": "/home/user/frontend",
      "log_level": "info",
      "env": [
        "PATH=/usr/local/bin:/usr/bin:/bin",
        "NODE_ENV=production",
        "API_KEY=secret123"
      ],
      "prompt": "Analyze the codebase and suggest improvements for performance",
      "token": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "ai_model": "claude-sonnet-4-20250514",
      "system_prompt": "You are a helpful coding assistant specializing in Typescript and React.",
      "exclude_tools": [
        "write",
        "shell(npm remove)",
        "shell(npm install)",
      ],
      "skills": [],
      "local_mcp_servers": {
        "filesystem": {
          "tools": ["read_file", "write_file", "list_directory"],
          "type": "stdio",
          "timeout": 30,
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/project"],
          "env": {
            "NODE_ENV": "production"
          },
          "cwd": "/home/user/project"
        }
      },
      "remote_mcp_servers": {},
      "timeout_sec": 300
    }
  ]
}
```

As you can see, you have a list of tasks under `tasks`, each having the following structure:

- **log_file**: Path where session logs will be written
- **cwd**: Current working directory for the copilot session
- **log_level**: Logging verbosity (e.g., "debug", "info", "warn", "error")
- **timeout_sec**: Maximum duration in seconds before the session times out
- **env**: Array of environment variables in `KEY=VALUE` format, available to tools and MCP servers
- **token**: GitHub personal access token for authentication. It is advised to use `$GITHUB_TOKEN` or `$GH_TOKEN` to reference environment variables, without pasting the actual token in the configuration file.
- **ai_model**: The AI model to use
- **system_prompt**: Custom instructions that define the AI's behavior and role
- **prompt**: The user's actual task or question for the AI to process
- **exclude_tools**: Black list of tools Copilot cannot use (e.g., `shell(rm)`, `write`, `shell(git push)`)
- **skills**: Directories containing files detailing higher-level capabilities or specialized knowledge areas the AI should employ
- **local_mcp_servers**: Stdio/local processes that run on the same machine:
  + **tools**: List of tool names this server provides
  + **type**: "stdio" or "local" - how to communicate with the server
  + **command**: Executable to run (e.g., "npx", "/usr/local/bin/tool")
  + **args**: Command-line arguments
  + **env**: Server-specific environment variables
  + **cwd**: Working directory for the server process
  + **timeout**: Maximum execution time in seconds
- **remote_mcp_servers**: HTTP/SSE endpoints for cloud-based tools:
  + **tools**: List of tool names this server provides
  + **type**: "http" or "sse" (Server-Sent Events for streaming)
  + **url**: API endpoint URL
  + **headers**: HTTP headers for authentication and content type
  + **timeout**: Maximum request duration in seconds

Take a look at the [example configuration](./multipilot.config.json) to see a real-world example on how you can use multipilot to run two tasks concurrently on two different projects (`multipilot` and [`workflows-acp`](https://github.com/AstraBert/workflows-acp)) to identify the underlying workflow engines that they are using.

Once the configuration is defined, run the tasks:

```bash
# use with a custom configuration path
multipilot --config config.json
# use with the default config, `multipilot.config.json`
multipilot
```

Each task will be run concurrently and, at the end, you will have a report of successfull and failed tasks.

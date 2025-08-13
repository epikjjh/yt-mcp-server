# Model Context Protocol (MCP) Implementation

## Overview

This document explains how our YouTube MCP Server implements the Model Context Protocol (MCP) to provide a suite of tools for interacting with YouTube. The MCP is a standardized protocol that enables AI assistants to securely connect to external data sources and tools.

## What is MCP?

The **Model Context Protocol (MCP)** is an open standard for connecting AI assistants to external systems. It provides:

- **Standardized Communication**: JSON-RPC 2.0 based messaging
- **Security**: OAuth 2.0 authentication and authorization
- **Tool Discovery**: Dynamic capability advertisement
- **Extensibility**: Plugin-like architecture for adding new capabilities.

## MCP Method Implementation

### 1. Initialize Method

**Purpose**: Establishes the MCP session and exchanges capabilities. The request and response are standard and not detailed here.

### 2. Tools/List Method

**Purpose**: Advertises available tools to the AI assistant.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "get_video_metadata",
        "description": "Gets detailed information for a specific video.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "video_id": {
              "type": "string",
              "description": "The ID of the YouTube video."
            }
          },
          "required": ["video_id"]
        }
      },
      {
        "name": "search_videos",
        "description": "Searches for YouTube videos.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "query": {
              "type": "string",
              "description": "The search term."
            },
            "limit": {
              "type": "integer",
              "description": "Maximum number of results."
            }
          },
          "required": ["query"]
        }
      },
      {
        "name": "get_video_comments",
        "description": "Fetches top-level comment threads for a video.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "video_id": {
              "type": "string",
              "description": "The ID of the YouTube video."
            }
          },
          "required": ["video_id"]
        }
      }
    ]
  }
}
```

### 3. Tools/Call Method

**Purpose**: Executes a specific tool with provided arguments.

**Request Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "get_video_metadata",
    "arguments": {
      "video_id": "kYB8IZa5AuE"
    }
  }
}
```

**Success Response Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "json",
        "json": {
          "title": "Google I/O 2023 Developer Keynote",
          "viewCount": "584K",
          "likeCount": "12K",
          "description": "Watch the full developer keynote from Google I/O 2023..."
        }
      }
    ]
  }
}
```

**Error Response Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Authentication required. Please visit /oauth/authorize to log in."
      }
    ],
    "isError": true
  }
}
```

## Authentication

This server uses a simplified, single-user OAuth 2.0 flow. Before calling any tools, the user must authenticate by visiting the `/oauth/authorize` endpoint in their browser. This is a one-time action per server session. If the server is restarted, the user must re-authenticate.

Each `tools/call` request is sent to the protected `/mcp` endpoint, and the server validates the user's in-memory session before proceeding. 
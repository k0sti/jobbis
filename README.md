# Finnish Job Market Integration Project

This repository contains tools and integrations for accessing the Finnish job market via Työmarkkinatori.fi.

## 📁 Project Structure

```
jobbis/
├── mcp/                    # Työmarkkinatori MCP Server
│   ├── src/               # Source code
│   ├── doc/               # Documentation
│   ├── tests/             # Test suite
│   └── README.md          # MCP Server documentation
└── README.md              # This file
```

## 🚀 Projects

### [Työmarkkinatori MCP Server](mcp/)

MCP (Model Context Protocol) server that enables AI assistants to search and retrieve job listings from the Finnish job market using the official Työmarkkinatori REST API.

**Status:** Co-created with Claude Code. Implementation complete but not yet tested.

**Features:**
- 🔐 OAuth 2.0 authentication
- 🌐 Työmarkkinatori REST API integration
- 🔍 Job search with multiple filters

**Quick Links:**
- [MCP Server Documentation](mcp/README.md)
- [Architecture & Design](mcp/doc/architecture.md)
- [API Reference](mcp/doc/api_reference.md)
- [API Credentials Setup Guide](mcp/doc/api_credentials_guide.md)
- [Getting Started](mcp/doc/getting_started.md)

## 📋 Prerequisites

### API Access Required

To use the MCP server, you need API credentials from KEHA Centre:

1. **Email:** tmt-rajapinnat.keha@ely-keskus.fi
2. **Provide:** Organization details and use case
3. **Receive:** OAuth 2.0 credentials (CLIENT_ID, CLIENT_SECRET, TENANT_ID)

See [API Credentials Setup Guide](mcp/doc/api_credentials_guide.md) for detailed instructions.

## 🏁 Quick Start

### 1. Navigate to MCP Server

```bash
cd mcp
```

### 2. Build

```bash
just build
# or: go build -o bin/tyomarkkinatori-mcp ./cmd/tyomarkkinatori-mcp
```

### 3. Run

```bash
./bin/tyomarkkinatori-mcp
```

### 4. Configure MCP

Add to Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "/absolute/path/to/mcp/bin/tyomarkkinatori-mcp",
      "env": {
        "CLIENT_ID": "your-client-id",
        "CLIENT_SECRET": "your-client-secret",
        "TENANT_ID": "your-tenant-id"
      }
    }
  }
}
```

See [MCP Server Documentation](mcp/README.md) for detailed configuration options.

## 📚 Documentation

### General
- [Project README](README.md) - This file
- [MCP Server Documentation](mcp/README.md) - Complete server documentation

### Setup & Configuration
- [API Credentials Guide](mcp/doc/api_credentials_guide.md) - How to get API access

### Technical Documentation
- [Architecture](mcp/doc/architecture.md) - System design and decisions
- [Implementation (Go)](mcp/doc/implementation-go.md) - Go implementation guide
- [Implementation (TypeScript)](mcp/doc/implementation-typescript.md) - TypeScript reference implementation
- [API Reference](mcp/doc/api_reference.md) - Tool specifications

## 🔧 MCP Tools Available

### search_jobs
Search for job listings with advanced filters:
- Keywords, location, ESCO occupation groups
- Employer type, working hours, duration
- Date filters, language selection
- Pagination support (100-500 results per page)

### get_job_details
Retrieve complete information for a specific job listing.

### list_categories
Get available filter options including:
- ESCO occupation groups
- Employer types
- Working hours and duration options
- Regions and languages

## 📊 API Details

### Endpoints
- **Production:** `https://integraatiot.tyomarkkinatori.fi`
- **QA/Test:** `https://integraatiot-qa.tyomarkkinatori.fi`

## 📄 License

MIT

## 📞 Support

### For Technical Issues
- `cd mcp`
- `cat README.md >` to your favourite coding LLM

### For API Access
- **Email:** tmt-rajapinnat.keha@ely-keskus.fi
- **Purpose:** Requesting credentials, API questions

## ⚠️ Important Notes

1. **API Registration Required:** This project requires official API credentials from KEHA Centre
2. **Test Environment:** Test first with QA environment
3. **Rate Limiting:** Respect API rate limits
4. **Compliance:** Follow Työmarkkinatori API terms of use and GDPR regulations

## 🔗 External Resources

- [Työmarkkinatori.fi](https://tyomarkkinatori.fi/) - Finnish job market portal
- [ESCO Classification](https://ec.europa.eu/esco/portal) - European Skills/Competences/Qualifications
- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP documentation
- [Claude Desktop](https://claude.ai/) - AI assistant

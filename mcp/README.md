# Työmarkkinatori MCP Server

MCP (Model Context Protocol) server for searching Finnish job market via [Työmarkkinatori.fi](https://tyomarkkinatori.fi/).

## Overview

This server enables AI assistants like Claude to search and retrieve job listings from the Finnish job market using the official Työmarkkinatori REST API, providing seamless integration for job search queries.

MCP server is co-created with Claude Code (v 2.0.8, 2025-10-15).
Seems legit, but has not been tested.

## Features

- 🔍 **Job Search** - Search jobs by keywords, location, category, and employment type
- 📄 **Job Details** - Retrieve full details for specific job listings
- 🏷️ **Categories** - List available job categories, employment types, and regions
- 🔐 **OAuth 2.0** - Secure authentication via Microsoft Identity Platform
- 🌐 **Official API** - Uses Työmarkkinatori REST API (requires registration)
- ⚡ **Caching** - Built-in caching for improved performance
- 🚦 **Rate Limiting** - Respectful API usage with rate limiting
- 🇫🇮 **Finnish Market** - Specialized for Finnish job market terminology
- 📊 **ESCO Classification** - European standard for skills and occupations

## Prerequisites

### API Access

**Important:** This server requires API credentials from KEHA Centre.

1. **Request API Access:**
   - Email: tmt-rajapinnat.keha@ely-keskus.fi
   - Provide organization details and use case
   - Accept Työmarkkinatori API terms of use

2. **Receive Credentials:**
   - You'll receive: `CLIENT_ID`, `CLIENT_SECRET`, `TENANT_ID`
   - First for QA environment, then production after testing

See [API Credentials Guide](doc/api_credentials_guide.md) for detailed registration process.

## Quick Start

### Installation

```bash
cd mcp
go mod download
```

### Configuration

1. Build:
```bash
just build
# or: go build -o bin/tyomarkkinatori-mcp ./cmd/tyomarkkinatori-mcp
```

2. Configure Claude Desktop (`claude_desktop_config.json`):

**Required Configuration:**
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

**Optional Environment Variables:**
| Variable | Default | Description |
|----------|---------|-------------|
| API_BASE_URL | `https://integraatiot.tyomarkkinatori.fi` | API endpoint (use `https://integraatiot-qa.tyomarkkinatori.fi` for testing) |
| CACHE_TTL_MS | `900000` | Cache TTL in milliseconds (15 minutes) |
| RATE_LIMIT_RPS | `2` | Requests per second |
| RATE_LIMIT_BURST | `10` | Maximum burst capacity |

**Full Configuration Example:**
```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "/absolute/path/to/mcp/bin/tyomarkkinatori-mcp",
      "env": {
        "CLIENT_ID": "your-client-id",
        "CLIENT_SECRET": "your-client-secret",
        "TENANT_ID": "your-tenant-id",
        "API_BASE_URL": "https://integraatiot.tyomarkkinatori.fi",
        "CACHE_TTL_MS": "900000",
        "RATE_LIMIT_RPS": "2",
        "RATE_LIMIT_BURST": "10"
      }
    }
  }
}
```

### Development

```bash
just run
# or: go run ./cmd/tyomarkkinatori-mcp
```

## Tools

### 1. search_jobs

Search for job listings with filters.

**Example:**
```json
{
  "query": "ohjelmistokehittäjä",
  "location": "Helsinki",
  "employment_type": "kokoaikainen",
  "limit": 20
}
```

### 2. get_job_details

Get detailed information about a specific job.

**Example:**
```json
{
  "job_id": "12345"
}
```

### 3. list_categories

List available categories, employment types, and regions.

**Example:**
```json
{}
```

## Documentation

- [Architecture & Design](doc/architecture.md) - System architecture and design decisions
- [Implementation (Go)](doc/implementation-go.md) - Go implementation guide
- [Implementation (TypeScript)](doc/implementation-typescript.md) - TypeScript reference implementation
- [API Reference](doc/api_reference.md) - Complete tool reference and examples

## Technology Stack

- **Language:** Go 1.21+
- **MCP SDK:** github.com/modelcontextprotocol/go-sdk
- **HTTP Client:** net/http (standard library)
- **OAuth:** golang.org/x/oauth2
- **Authentication:** OAuth 2.0 (Microsoft Identity Platform)
- **API:** Työmarkkinatori REST API

## Project Structure

```
mcp/
├── cmd/
│   └── tyomarkkinatori-mcp/
│       └── main.go           # Entry point
├── internal/
│   ├── server/               # MCP server
│   ├── tools/                # Tool handlers
│   ├── auth/                 # OAuth 2.0 authentication
│   ├── client/               # REST API client
│   ├── models/               # Data models
│   ├── utils/                # Utilities
│   └── config/               # Configuration
├── doc/                      # Documentation
├── go.mod                    # Go module definition
└── justfile                  # Build automation
```

## Usage Examples

### Finding Recent Developer Jobs

```
User: "Find recent software developer jobs in Helsinki"

Assistant uses: search_jobs with:
{
  "query": "ohjelmistokehittäjä",
  "location": "Helsinki",
  "published_after": "2024-03-01"
}
```

### Getting Job Details

```
User: "Tell me more about job 12345"

Assistant uses: get_job_details with:
{
  "job_id": "12345"
}
```

## Finnish Terminology

- **kokoaikainen** - Full-time
- **osa-aikainen** - Part-time
- **määräaikainen** - Fixed-term
- **kehittäjä** - Developer
- **hoitaja** - Nurse
- **insinööri** - Engineer

See [API Reference](doc/api_reference.md) for complete terminology guide.

## Configuration

### Required Environment Variables

```bash
CLIENT_ID=your-client-id              # From KEHA Centre
CLIENT_SECRET=your-client-secret      # From KEHA Centre
TENANT_ID=your-tenant-id              # From KEHA Centre
```

### Optional Environment Variables

```bash
# API endpoint (use QA for testing)
API_BASE_URL=https://integraatiot.tyomarkkinatori.fi
# QA: https://integraatiot-qa.tyomarkkinatori.fi

# Cache TTL in milliseconds (default: 15 minutes)
CACHE_TTL_MS=900000

# Rate limiting (default: 2 requests/second)
RATE_LIMIT_RPS=2
RATE_LIMIT_BURST=10
```

## API Environments

- **Production:** `https://integraatiot.tyomarkkinatori.fi`
- **QA/Test:** `https://integraatiot-qa.tyomarkkinatori.fi`

Always test with QA environment before using production.

## Rate Limiting

- Default: 2 requests per second
- Burst capacity: 10 requests
- Caching: 15 minutes for searches, 1 hour for details
- Respects API usage quotas

## Error Handling

The server provides informative error messages:

- OAuth authentication errors
- Parameter validation errors
- Network connectivity issues
- API response errors
- Rate limiting

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Maintenance

### API Changes

Monitor Työmarkkinatori API for changes:

1. Subscribe to API changelog (if available)
2. Test regularly with QA environment
3. Update API client when endpoints change
4. Document changes

### Credential Management

- Store credentials in MCP client config
- Rotate credentials periodically
- Monitor OAuth token expiration

### Dependencies

Keep dependencies updated:

```bash
just update
# or: go get -u ./... && go mod tidy
```

## License

MIT

## Support

For issues and questions:
- Check documentation in `doc/` folder
- Review error messages carefully
- Test with simpler queries first

## Roadmap

- [ ] Job alerts and notifications
- [ ] Application tracking via import API
- [ ] Resume matching with ESCO skills
- [ ] Multi-language support (Swedish, English via API)
- [ ] Integration with other job boards
- [ ] Analytics and job market trends
- [ ] Advanced ESCO classification features

## Acknowledgments

Built with:
- [Model Context Protocol (MCP)](https://github.com/modelcontextprotocol)
- [Työmarkkinatori REST API](https://tyomarkkinatori.fi/)
- [Axios](https://axios-http.com/) for HTTP requests
- Microsoft Identity Platform for OAuth 2.0

## Security

- OAuth 2.0 Client Credentials Flow
- Credentials stored in environment variables
- Tokens cached in memory only
- No sensitive data in logs
- HTTPS-only communication

## Disclaimer

This tool uses the official Työmarkkinatori API. API access requires registration with KEHA Centre. Please comply with Työmarkkinatori's API terms of use and applicable data protection regulations.

**Vibed with Claude**

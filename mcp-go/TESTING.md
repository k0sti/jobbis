# Testing with Työmarkkinatori API

## Getting Test Credentials

Contact the Työmarkkinatori API team to obtain test environment credentials:
- `CLIENT_ID` - Your unique test client identifier
- `CLIENT_SECRET` - Your secret key for authentication

## Environment Configuration

The MCP server automatically configures itself based on the `ENVIRONMENT` variable:

### Test Environment (Recommended for Development)
```bash
export CLIENT_ID="your-test-client-id"
export CLIENT_SECRET="your-test-client-secret"
export ENVIRONMENT="test"
```

This automatically uses:
- **API**: `https://integraatiot-qa.tyomarkkinatori.fi`
- **Token URL**: `https://tedigidevb2c.b2clogin.com/tedigidevb2c.onmicrosoft.com/B2C_1A_SIGNIN/oauth2/v2.0/token`
- **Scope**: `https://tedigidevb2c.onmicrosoft.com/cd93ea6e-c118-4100-b2bd-5676e1ea4c50/.default`

### Production Environment
```bash
export CLIENT_ID="your-production-client-id"
export CLIENT_SECRET="your-production-client-secret"
export ENVIRONMENT="production"
```

This automatically uses:
- **API**: `https://integraatiot.tyomarkkinatori.fi`
- **Token URL**: `https://tedigib2c.b2clogin.com/tedigib2c.onmicrosoft.com/B2C_1A_SIGNIN/oauth2/v2.0/token`
- **Scope**: `https://tedigib2c.onmicrosoft.com/e9343614-0f96-437e-9f5e-1d392e756675/.default`

## Testing Methods

### 1. Using MCP Inspector (Interactive GUI)

The easiest way to test:

```bash
# Set test credentials
export CLIENT_ID="your-test-client-id"
export CLIENT_SECRET="your-test-client-secret"
export ENVIRONMENT="test"

# Run with MCP Inspector
npx @modelcontextprotocol/inspector ./bin/tyomarkkinatori-mcp
```

This opens a web interface at `http://localhost:5173` where you can:
- View available tools
- Call tools with custom parameters
- See request/response data
- Debug authentication issues

### 2. Using Claude Code

Update `.claude/mcp.json` with your test credentials:

```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "/home/k0/work/jobbis/mcp-go/bin/tyomarkkinatori-mcp",
      "env": {
        "CLIENT_ID": "your-actual-test-client-id",
        "CLIENT_SECRET": "your-actual-test-client-secret",
        "ENVIRONMENT": "test"
      }
    }
  }
}
```

Then restart Claude Code and use these tools:
- `mcp__tyomarkkinatori__search_jobs` - Search for jobs
- `mcp__tyomarkkinatori__get_job_details` - Get job details
- `mcp__tyomarkkinatori__list_categories` - List available categories

### 3. Manual Testing with curl

Get an access token:

```bash
curl -X POST "https://tedigidevb2c.b2clogin.com/tedigidevb2c.onmicrosoft.com/B2C_1A_SIGNIN/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=https://tedigidevb2c.onmicrosoft.com/cd93ea6e-c118-4100-b2bd-5676e1ea4c50/.default" \
  -d "grant_type=client_credentials"
```

Then test the API:

```bash
# Search for jobs
curl -X GET "https://integraatiot-qa.tyomarkkinatori.fi/jobpostingprovider/v1/tyopaikat?hakusana=ohjelmistokehittäjä" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Accept: application/json"
```

## Running Unit Tests

All tests use mocks and don't require credentials:

```bash
# Run all tests
just test

# Or using go directly
go test ./... -v

# With coverage
go test ./... -cover
```

## Example Test Scenarios

### Search for Software Developer Jobs in Helsinki

```json
{
  "query": "ohjelmistokehittäjä",
  "location": "Helsinki",
  "working_hours": "full-time",
  "language": "fi"
}
```

### Get Job Details

```json
{
  "job_id": "some-job-id-from-search-results"
}
```

### List Available Categories

```json
{}
```

## Troubleshooting

### Authentication Errors

If you see authentication errors:
1. Verify your `CLIENT_ID` and `CLIENT_SECRET` are correct
2. Check that `ENVIRONMENT` is set to "test" for test credentials
3. Ensure the credentials haven't expired (contact API team)

### API Errors

If you see API errors:
1. Check the API is accessible: `curl https://integraatiot-qa.tyomarkkinatori.fi/health`
2. Verify your token is valid using the curl command above
3. Check rate limits (default: 2 requests/second)

### Log Output

The server logs OAuth token requests. If authentication succeeds you'll see:
```
Requesting OAuth access token
OAuth token obtained successfully
```

If it fails:
```
Failed to obtain OAuth token: <error details>
```

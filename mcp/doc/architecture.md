# Työmarkkinatori MCP Server - Architecture & Design

## Overview

This MCP (Model Context Protocol) server provides access to the Finnish job market through Työmarkkinatori.fi, enabling AI assistants to search and retrieve job listings for users.

## Architecture

### High-Level Components

```
┌─────────────────────────────────────────┐
│         MCP Client (Claude)             │
└────────────────┬────────────────────────┘
                 │ MCP Protocol
                 │
┌────────────────▼────────────────────────┐
│      Työmarkkinatori MCP Server         │
│  ┌───────────────────────────────────┐  │
│  │     Tool Handlers                 │  │
│  │  - search_jobs                    │  │
│  │  - get_job_details                │  │
│  │  - list_categories                │  │
│  └──────────────┬────────────────────┘  │
│                 │                        │
│  ┌──────────────▼────────────────────┐  │
│  │  OAuth 2.0 Authentication         │  │
│  │  - Token management               │  │
│  │  - Credential storage             │  │
│  │  - Token refresh                  │  │
│  └──────────────┬────────────────────┘  │
│                 │                        │
│  ┌──────────────▼────────────────────┐  │
│  │     REST API Client Layer         │  │
│  │  - HTTP client (axios)            │  │
│  │  - Request builder                │  │
│  │  - Response parser (JSON)         │  │
│  └──────────────┬────────────────────┘  │
│                 │                        │
│  ┌──────────────▼────────────────────┐  │
│  │     Data Models                   │  │
│  │  - JobListing                     │  │
│  │  - SearchParams                   │  │
│  │  - FilterOptions                  │  │
│  └───────────────────────────────────┘  │
└────────────────┬────────────────────────┘
                 │ HTTPS + OAuth 2.0
                 │
┌────────────────▼────────────────────────┐
│   Työmarkkinatori REST API              │
│   - integraatiot.tyomarkkinatori.fi     │
│   - OAuth 2.0 (MS Identity Platform)    │
└─────────────────────────────────────────┘
```

## Core Components

### 1. MCP Server Core

**Responsibilities:**
- Handle MCP protocol communication
- Register and expose tools
- Manage server lifecycle
- Handle errors and logging

**Technology:**
- Node.js/TypeScript
- MCP SDK (@modelcontextprotocol/sdk)

### 2. Tool Handlers

#### search_jobs Tool
**Purpose:** Search for job listings based on various criteria using the official API

**Parameters:**
- `query` (string, optional): Keyword search (job title, description, skills)
- `location` (string, optional): Location/region filter
- `occupation_group` (string, optional): ESCO occupation group code
- `employer_type` (string, optional): company, public, nonprofit
- `working_hours` (string, optional): full-time, part-time
- `duration` (string, optional): permanent, temporary
- `published_after` (string, optional): ISO date string
- `language` (string, optional): fi, sv, en (default: fi)
- `page` (number, optional): Page number for pagination (default: 0)
- `page_size` (number, optional): Results per page, 100-500 (default: 100)

**Returns:** Array of job listings with summary information

#### get_job_details Tool
**Purpose:** Retrieve full details for a specific job listing

**Parameters:**
- `job_id` (string, required): Unique identifier for the job

**Returns:** Complete job listing information

#### list_categories Tool
**Purpose:** Get available job categories and filters

**Parameters:** None

**Returns:** List of available categories, employment types, and regions

### 3. OAuth 2.0 Authentication Layer

**Responsibilities:**
- Manage OAuth 2.0 Client Credentials Flow
- Store and refresh access tokens
- Handle authentication errors
- Secure credential storage

**Authentication Flow:**
1. Load client credentials (client_id, client_secret, tenant_id)
2. Request access token from Microsoft Identity Platform
3. Cache token until expiry
4. Automatically refresh when expired
5. Include bearer token in all API requests

**Technology:**
- Microsoft Identity Platform OAuth 2.0
- Client Credentials Flow
- Environment variables for credentials

### 4. REST API Client Layer

**Responsibilities:**
- Make authenticated HTTP requests to Työmarkkinatori REST API
- Build API requests with proper parameters
- Implement rate limiting
- Cache responses when appropriate
- Parse JSON responses and normalize data

**Key Features:**
- OAuth 2.0 token injection
- Retry logic with exponential backoff
- Request timeout handling
- Error mapping and user-friendly messages
- Response validation
- ESCO classification support

### 5. Data Models

#### JobListing
```typescript
interface JobListing {
  id: string;
  title: string;
  employer: string;
  location: string;
  employment_type: string;
  published_date: string;
  application_deadline?: string;
  summary: string;
  description?: string; // Full description (only in details)
  requirements?: string[];
  benefits?: string[];
  salary_info?: string;
  url: string;
}
```

#### SearchParams
```typescript
interface SearchParams {
  // Search & Basic Filters
  query?: string;
  location?: string;
  occupation_group?: string;

  // Employment Filters
  employer_type?: "company" | "public" | "nonprofit";
  working_hours?: "full-time" | "part-time";
  duration?: "permanent" | "temporary";

  // Date & Language
  published_after?: string;
  language?: "fi" | "sv" | "en";

  // Pagination
  page?: number;
  page_size?: number;
}
```

## Design Decisions

### 1. Official REST API vs Web Scraping
**Decision:** Use official Työmarkkinatori REST API

**Rationale:**
- Työmarkkinatori provides an official REST API for job retrieval
- More reliable than web scraping (no HTML structure changes)
- Better performance and structured JSON data
- ESCO classification support for European compatibility
- Proper authentication and rate limiting
- Requires registration with KEHA Centre

**API Details:**
- **Endpoints:**
  - Production: `https://integraatiot.tyomarkkinatori.fi/`
  - QA/Test: `https://integraatiot-qa.tyomarkkinatori.fi/`
- **Authentication:** OAuth 2.0 via Microsoft Identity Platform
- **Data Format:** JSON with ESCO classifications
- **Registration:** Contact tmt-rajapinnat.keha@ely-keskus.fi

### 2. OAuth 2.0 Authentication
**Decision:** Implement Client Credentials Flow

**Rationale:**
- Required by Työmarkkinatori API
- Secure service-to-service authentication
- Standard OAuth 2.0 protocol
- Token caching reduces authentication overhead
- Automatic token refresh

### 3. Caching Strategy
**Decision:** Implement in-memory caching with TTL

**Rationale:**
- Job listings don't change frequently (15-60 minute cache)
- Reduces load on target website
- Improves response times
- Categories/filters can be cached longer (24 hours)

### 4. Error Handling
**Decision:** Graceful degradation with informative errors

**Rationale:**
- Network issues are common
- Website structure may change
- Users need actionable error messages

### 5. Rate Limiting
**Decision:** Implement client-side rate limiting with API-compliant limits

**Rationale:**
- Respect API usage limits
- Prevent quota exhaustion
- Smooth request distribution
- Default: Conservative limits (adjust based on API documentation)
- Pagination support (100-500 items per page)

## Data Flow

### Job Search Flow
```
1. Client calls search_jobs tool with parameters
2. Server validates parameters
3. Check cache for recent results
4. If cache miss:
   a. Check OAuth token validity (refresh if expired)
   b. Build API request with search parameters
   c. Apply rate limiting
   d. Make authenticated GET request to API
   e. Parse JSON response
   f. Transform API data to JobListing format
   g. Cache results
5. Return results to client
```

### Authentication Flow
```
1. Server starts
2. Load credentials from environment variables
   - CLIENT_ID
   - CLIENT_SECRET
   - TENANT_ID
3. Request access token from Microsoft Identity:
   POST https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/token
   Body: client_id, client_secret, scope, grant_type
4. Store token with expiry time_ms
5. For each API request:
   - Check if token expired
   - If expired, refresh token (repeat step 3)
   - Add Authorization: Bearer {token} header
6. Handle authentication errors:
   - Invalid credentials → Log error and exit
   - Expired token → Refresh automatically
   - Network error → Retry with backoff
```

## Security Considerations

1. **Credential Management:**
   - Store API credentials in environment variables only
   - Never commit credentials to version control
   - Use `.env` files for local development
   - Secure credential storage in production

2. **Token Security:**
   - Store access tokens in memory only (no disk)
   - Clear tokens on server shutdown
   - Never log full tokens (only first/last chars)
   - Implement token refresh before expiry

3. **Input Validation:**
   - Sanitize all user inputs to prevent injection
   - Validate parameter types and ranges
   - Whitelist allowed values where possible

4. **Network Security:**
   - HTTPS only for all requests
   - Verify SSL certificates
   - Timeout all requests (10-30 seconds)

5. **Error Handling:**
   - Don't expose credentials in error messages
   - Sanitize API errors before returning to user
   - Log full errors securely for debugging

6. **Rate Limiting:**
   - Prevent quota exhaustion
   - Implement per-user limits if multi-tenant

## Scalability Considerations

1. **Caching:** Reduce external requests
2. **Async/Await:** Non-blocking I/O operations
3. **Connection Pooling:** Reuse HTTP connections
4. **Batch Processing:** Support multiple searches efficiently

## Monitoring & Logging

1. **Request Logging:** Track all API calls
2. **Error Logging:** Detailed error information for debugging
3. **Performance Metrics:** Response times, cache hit rates
4. **Health Checks:** Verify service availability

## API Registration Process

### Prerequisites
1. Organization must have a legitimate use case for accessing job data
2. Accept Työmarkkinatori API terms of use
3. Provide organization details and use case description

### Steps
1. **Submit Activation Form:** Contact KEHA Centre
2. **Email:** tmt-rajapinnat.keha@ely-keskus.fi
3. **Receive QA Credentials:** For testing environment
4. **Test Integration:** Verify implementation in QA
5. **Request Production Access:** After successful QA testing
6. **Receive Production Credentials:** client_id, client_secret, tenant_id

### Environments
- **QA/Test:** `https://integraatiot-qa.tyomarkkinatori.fi/`
- **Production:** `https://integraatiot.tyomarkkinatori.fi/`

## API Capabilities

### Supported Filters
- **Occupation Groups (ammattiryhmä):** ESCO classification
- **Employer Type (työllistäjä):** Public/private sector
- **Employment Duration (työn jatkuvuus):** Permanent/temporary
- **Working Hours (työaika):** Full-time/part-time
- **Language (kieli):** fi, sv, en
- **Location:** Country, region (maakunta), municipality (kunta)
- **Publication Date:** Filter by date range

### Pagination
- **Page Size:** 100-500 items per page
- **Parameters:** `sivu` (page number), `maara` (page size)

### Data Standards
- **ESCO Classification:** European Skills, Competences, Qualifications
- **Ensures:** Compatibility with EU job mobility standards

## Future Enhancements

1. **Job Alerts:** Subscribe to new jobs matching criteria (API webhooks if available)
2. **Application Tracking:** Use import API for application management
3. **Resume Matching:** Match resumes to job requirements using ESCO skills
4. **Analytics:** Job market trends and statistics from aggregated data
5. **Multi-language Support:** Leverage API's fi, sv, en support
6. **Integration with Other Job Boards:** Expand beyond Työmarkkinatori
7. **Advanced ESCO Integration:** Full skill-based job matching
8. **Employer Integration:** Use import API for posting jobs (if applicable)

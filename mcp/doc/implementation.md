# Työmarkkinatori MCP Server - Implementation Specification

## Project Setup

### Technology Stack

- **Runtime:** Node.js 18+ / Bun
- **Language:** TypeScript 5+
- **MCP SDK:** @modelcontextprotocol/sdk
- **HTTP Client:** axios
- **OAuth Library:** @azure/msal-node (Microsoft Authentication Library)
- **Environment:** dotenv (for credential management)
- **Build Tool:** esbuild or tsc
- **Testing:** vitest or jest
- **Linting:** eslint + prettier

### Project Structure

```
mcp/
├── src/
│   ├── index.ts                 # Main entry point
│   ├── server.ts                # MCP server setup
│   ├── tools/
│   │   ├── search_jobs.ts       # Search jobs tool handler
│   │   ├── get_job_details.ts   # Job details tool handler
│   │   └── list_categories.ts   # Categories tool handler
│   ├── auth/
│   │   ├── oauth_client.ts      # OAuth 2.0 authentication
│   │   └── token_manager.ts     # Token storage and refresh
│   ├── client/
│   │   ├── api_client.ts        # REST API client
│   │   └── parser.ts            # JSON response parsing
│   ├── models/
│   │   ├── job.ts               # JobListing interface
│   │   ├── search.ts            # SearchParams interface
│   │   ├── filters.ts           # FilterOptions interface
│   │   └── api_types.ts         # API response types
│   ├── utils/
│   │   ├── cache.ts             # Caching implementation
│   │   ├── rate_limiter.ts      # Rate limiting
│   │   ├── validators.ts        # Input validation
│   │   └── logger.ts            # Logging utility
│   └── config.ts                # Configuration
├── tests/
│   ├── tools/
│   ├── client/
│   └── utils/
├── doc/                         # Documentation
├── package.json
├── tsconfig.json
├── .env.example
└── README.md
```

## Implementation Details

### 1. Main Entry Point (index.ts)

```typescript
#!/usr/bin/env node
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { TyomarkkinatoriServer } from "./server.js";

async function main() {
  const server = new TyomarkkinatoriServer();

  const transport = new StdioServerTransport();
  await server.connect(transport);

  console.error("Työmarkkinatori MCP server running on stdio");
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
```

### 2. Server Setup (server.ts)

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { searchJobsTool } from "./tools/search_jobs.js";
import { getJobDetailsTool } from "./tools/get_job_details.js";
import { listCategoriesTool } from "./tools/list_categories.js";

export class TyomarkkinatoriServer {
  private server: Server;

  constructor() {
    this.server = new Server(
      {
        name: "tyomarkkinatori-mcp-server",
        version: "1.0.0",
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  private setupHandlers() {
    // List available tools
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [
        {
          name: "search_jobs",
          description: "Search for job listings on Työmarkkinatori.fi",
          inputSchema: {
            type: "object",
            properties: {
              query: {
                type: "string",
                description: "Keywords to search (job title, skills, description)",
              },
              location: {
                type: "string",
                description: "Location or region (e.g., Helsinki, Tampere)",
              },
              category: {
                type: "string",
                description: "Job category or industry",
              },
              employment_type: {
                type: "string",
                description: "Employment type (kokoaikainen, osa-aikainen, määräaikainen)",
              },
              published_after: {
                type: "string",
                description: "ISO date string (e.g., 2024-01-01)",
              },
              limit: {
                type: "number",
                description: "Maximum results to return (default: 20, max: 100)",
                default: 20,
              },
            },
          },
        },
        {
          name: "get_job_details",
          description: "Get detailed information about a specific job listing",
          inputSchema: {
            type: "object",
            properties: {
              job_id: {
                type: "string",
                description: "Unique identifier for the job listing",
              },
            },
            required: ["job_id"],
          },
        },
        {
          name: "list_categories",
          description: "Get available job categories, employment types, and regions",
          inputSchema: {
            type: "object",
            properties: {},
          },
        },
      ],
    }));

    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      switch (name) {
        case "search_jobs":
          return await searchJobsTool(args);
        case "get_job_details":
          return await getJobDetailsTool(args);
        case "list_categories":
          return await listCategoriesTool(args);
        default:
          throw new Error(`Unknown tool: ${name}`);
      }
    });
  }

  async connect(transport: any) {
    await this.server.connect(transport);
  }
}
```

### 3. OAuth 2.0 Authentication (auth/oauth_client.ts)

```typescript
import axios from "axios";
import { logger } from "../utils/logger.js";
import { config } from "../config.js";

interface TokenResponse {
  access_token: string;
  expires_in: number;
  token_type: string;
}

export class OAuthClient {
  private accessToken: string | null = null;
  private tokenExpiry_ms: number = 0;

  async getAccessToken(): Promise<string> {
    // Check if we have a valid token
    if (this.accessToken && Date.now() < this.tokenExpiry_ms) {
      return this.accessToken;
    }

    // Request new token
    return await this.requestNewToken();
  }

  private async requestNewToken(): Promise<string> {
    const tokenUrl = `https://login.microsoftonline.com/${config.tenantId}/oauth2/v2.0/token`;

    const params = new URLSearchParams({
      client_id: config.clientId,
      client_secret: config.clientSecret,
      scope: "https://graph.microsoft.com/.default",
      grant_type: "client_credentials",
    });

    try {
      logger.info("Requesting OAuth access token");

      const response = await axios.post<TokenResponse>(tokenUrl, params, {
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
      });

      this.accessToken = response.data.access_token;
      // Set expiry with 5 minute buffer
      this.tokenExpiry_ms = Date.now() + (response.data.expires_in - 300) * 1000;

      logger.info("OAuth token obtained successfully");
      return this.accessToken;
    } catch (error) {
      logger.error("Failed to obtain OAuth token", error);
      throw new Error("Authentication failed - check credentials");
    }
  }

  clearToken(): void {
    this.accessToken = null;
    this.tokenExpiry_ms = 0;
  }
}
```

### 4. Configuration (config.ts)

```typescript
import * as dotenv from "dotenv";

dotenv.config();

export const config = {
  // OAuth 2.0 credentials
  clientId: process.env.CLIENT_ID || "",
  clientSecret: process.env.CLIENT_SECRET || "",
  tenantId: process.env.TENANT_ID || "",

  // API endpoints
  apiBaseUrl: process.env.API_BASE_URL || "https://integraatiot.tyomarkkinatori.fi",

  // Environment
  environment: process.env.NODE_ENV || "production",

  // Cache settings
  cacheTTL_ms: parseInt(process.env.CACHE_TTL_MS || "900000"), // 15 minutes

  // Rate limiting
  rateLimit_rps: parseFloat(process.env.RATE_LIMIT_RPS || "2"), // 2 requests per second
  rateLimitBurst: parseInt(process.env.RATE_LIMIT_BURST || "10"),
};

// Validate required config
if (!config.clientId || !config.clientSecret || !config.tenantId) {
  console.error("ERROR: Missing required environment variables:");
  if (!config.clientId) console.error("  - CLIENT_ID");
  if (!config.clientSecret) console.error("  - CLIENT_SECRET");
  if (!config.tenantId) console.error("  - TENANT_ID");
  console.error("\nPlease set these in your .env file");
  process.exit(1);
}
```

### 5. REST API Client (client/api_client.ts)

```typescript
import axios, { AxiosInstance } from "axios";
import { OAuthClient } from "../auth/oauth_client.js";
import { RateLimiter } from "../utils/rate_limiter.js";
import { Cache } from "../utils/cache.js";
import { logger } from "../utils/logger.js";
import { config } from "../config.js";
import { SearchParams, JobListing } from "../models/job.js";

export class TyomarkkinatoriClient {
  private client: AxiosInstance;
  private oauthClient: OAuthClient;
  private rateLimiter: RateLimiter;
  private cache: Cache;

  constructor() {
    this.oauthClient = new OAuthClient();

    this.client = axios.create({
      baseURL: `${config.apiBaseUrl}/jobpostingprovider/v1`,
      timeout: 30000,
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
    });

    this.rateLimiter = new RateLimiter({
      requestsPerSecond: config.rateLimit_rps,
      maxBurst: config.rateLimitBurst,
    });

    this.cache = new Cache({
      defaultTTL: config.cacheTTL_ms,
    });
  }

  async searchJobs(params: SearchParams): Promise<JobListing[]> {
    const cacheKey = `search:${JSON.stringify(params)}`;
    const cached = this.cache.get<JobListing[]>(cacheKey);

    if (cached) {
      logger.info("Cache hit for search");
      return cached;
    }

    await this.rateLimiter.waitForToken();

    try {
      const token = await this.oauthClient.getAccessToken();

      const response = await this.client.get("/tyopaikat", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        params: this.buildAPIParams(params),
      });

      const jobs = this.parseJobListings(response.data);
      this.cache.set(cacheKey, jobs);

      return jobs;
    } catch (error) {
      logger.error("Search request failed", error);
      throw new Error("Failed to search jobs");
    }
  }

  async getJobDetails(jobId: string): Promise<JobListing> {
    const cacheKey = `job:${jobId}`;
    const cached = this.cache.get(cacheKey);

    if (cached) {
      return cached;
    }

    await this.rateLimiter.waitForToken();

    try {
      const response = await this.client.get(`/tyopaikka/${jobId}`);
      const job = this.parseJobDetails(response.data);
      this.cache.set(cacheKey, job, 3600000); // 1 hour

      return job;
    } catch (error) {
      logger.error("Job details request failed", error);
      throw new Error(`Failed to fetch job details for ID: ${jobId}`);
    }
  }

  private buildAPIParams(params: SearchParams): Record<string, any> {
    // Transform our params to match the API's query format
    const apiParams: Record<string, any> = {
      sivu: params.page || 0,          // Page number (0-based)
      maara: params.page_size || 100,  // Results per page (100-500)
      kieli: params.language || "fi",  // Language (fi, sv, en)
    };

    // Add search query if provided
    if (params.query) {
      apiParams.hakusana = params.query;
    }

    // Add location filter
    if (params.location) {
      apiParams.sijainti = params.location;
    }

    // Add ESCO occupation group
    if (params.occupation_group) {
      apiParams.ammattiryhmä = params.occupation_group;
    }

    // Map employer type to API codes
    if (params.employer_type) {
      const employerTypeMap: Record<string, string> = {
        "company": "01",
        "public": "02",
        "nonprofit": "03",
      };
      apiParams.tyollistaja = employerTypeMap[params.employer_type];
    }

    // Map working hours to API codes
    if (params.working_hours) {
      const workingHoursMap: Record<string, string> = {
        "full-time": "01",
        "part-time": "02",
      };
      apiParams.työaika = workingHoursMap[params.working_hours];
    }

    // Map duration to API codes
    if (params.duration) {
      const durationMap: Record<string, string> = {
        "permanent": "01",
        "temporary": "02",
      };
      apiParams["työn jatkuvuus"] = durationMap[params.duration];
    }

    // Add date filter if provided
    if (params.published_after) {
      apiParams.julkaisupvm = params.published_after;
    }

    return apiParams;
  }

  private parseJobListings(data: any): JobListing[] {
    const jobs: JobListing[] = [];

    // API returns array in 'tulokset' field
    const results = data.tulokset || data || [];

    for (const item of results) {
      jobs.push({
        id: item.id || item.ilmoitusId || "",
        title: item.otsikko || item.ammattinimike || "",
        employer: item.tyonantaja || "",
        location: this.formatLocation(item.sijainti || item.tyopaikka),
        employment_type: this.formatEmploymentType(item.tyosuhteenLaji || item.tyonLaji),
        published_date: item.julkaisupvm || item.luontipvm || new Date().toISOString(),
        application_deadline: item.hakuaika?.paattyy || undefined,
        summary: item.kuvaus || item.tehtavakuvaus || "",
        url: `https://tyomarkkinatori.fi/tyopaikka/${item.id || item.ilmoitusId}`,
      });
    }

    return jobs;
  }

  private formatLocation(location: any): string {
    if (typeof location === "string") return location;
    if (location?.kunta) return location.kunta;
    if (location?.maakunta) return location.maakunta;
    return "Finland";
  }

  private formatEmploymentType(type: any): string {
    const typeMap: Record<string, string> = {
      "01": "kokoaikainen",
      "02": "osa-aikainen",
      "03": "määräaikainen",
    };
    return typeMap[type] || type || "kokoaikainen";
  }
}
```

### 7. Tool Implementation (tools/search_jobs.ts)

```typescript
import { TyomarkkinatoriClient } from "../client/api_client.js";
import { validateSearchParams } from "../utils/validators.js";
import { logger } from "../utils/logger.js";

const client = new TyomarkkinatoriClient();

export async function searchJobsTool(args: any) {
  try {
    const params = validateSearchParams(args);
    logger.info("Searching jobs", params);

    const jobs = await client.searchJobs(params);

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(
            {
              total: jobs.length,
              jobs: jobs,
            },
            null,
            2
          ),
        },
      ],
    };
  } catch (error) {
    logger.error("Search jobs tool error", error);
    return {
      content: [
        {
          type: "text",
          text: `Error searching jobs: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}
```

### 6. Caching (utils/cache.ts)

```typescript
interface CacheEntry<T> {
  data: T;
  expiry_ms: number;
}

export class Cache {
  private store: Map<string, CacheEntry<any>>;
  private defaultTTL_ms: number;

  constructor(options: { defaultTTL: number }) {
    this.store = new Map();
    this.defaultTTL_ms = options.defaultTTL;

    // Cleanup expired entries every minute
    setInterval(() => this.cleanup(), 60000);
  }

  get<T>(key: string): T | null {
    const entry = this.store.get(key);

    if (!entry) return null;

    if (Date.now() > entry.expiry_ms) {
      this.store.delete(key);
      return null;
    }

    return entry.data;
  }

  set<T>(key: string, data: T, ttl_ms?: number): void {
    const expiry_ms = Date.now() + (ttl_ms || this.defaultTTL_ms);
    this.store.set(key, { data, expiry_ms });
  }

  delete(key: string): void {
    this.store.delete(key);
  }

  clear(): void {
    this.store.clear();
  }

  private cleanup(): void {
    const now_ms = Date.now();
    for (const [key, entry] of this.store.entries()) {
      if (now_ms > entry.expiry_ms) {
        this.store.delete(key);
      }
    }
  }
}
```

### 7. Rate Limiter (utils/rate_limiter.ts)

```typescript
export class RateLimiter {
  private tokens: number;
  private maxTokens: number;
  private refillRate_ms: number;
  private lastRefill_ms: number;

  constructor(options: { requestsPerSecond: number; maxBurst: number }) {
    this.maxTokens = options.maxBurst;
    this.tokens = this.maxTokens;
    this.refillRate_ms = 1000 / options.requestsPerSecond;
    this.lastRefill_ms = Date.now();
  }

  async waitForToken(): Promise<void> {
    this.refillTokens();

    if (this.tokens >= 1) {
      this.tokens -= 1;
      return;
    }

    const waitTime_ms = this.refillRate_ms;
    await new Promise((resolve) => setTimeout(resolve, waitTime_ms));

    return this.waitForToken();
  }

  private refillTokens(): void {
    const now_ms = Date.now();
    const timePassed_ms = now_ms - this.lastRefill_ms;
    const tokensToAdd = timePassed_ms / this.refillRate_ms;

    this.tokens = Math.min(this.maxTokens, this.tokens + tokensToAdd);
    this.lastRefill_ms = now_ms;
  }
}
```

## Data Models

### models/job.ts

```typescript
export interface JobListing {
  // Basic information
  id: string;
  title: string;
  employer: string;
  location: string;
  employment_type: string;

  // Dates
  published_date: string;              // ISO 8601 format
  application_deadline?: string;       // ISO 8601 format (optional)

  // Description
  summary: string;
  description?: string;                // Full description (only in details)

  // Additional details
  requirements?: string[];
  benefits?: string[];
  salary_info?: string;

  // Link
  url: string;
}

export interface SearchParams {
  // Search & Basic Filters
  query?: string;              // Keyword search → API: hakusana
  location?: string;           // Location filter → API: sijainti
  occupation_group?: string;   // ESCO occupation group → API: ammattiryhmä

  // Employment Filters
  employer_type?: "company" | "public" | "nonprofit";  // → API: tyollistaja
  working_hours?: "full-time" | "part-time";           // → API: työaika
  duration?: "permanent" | "temporary";                // → API: työn jatkuvuus

  // Date & Language
  published_after?: string;    // ISO 8601 → API: julkaisupvm
  language?: "fi" | "sv" | "en";  // → API: kieli

  // Pagination
  page?: number;               // 0-based → API: sivu
  page_size?: number;          // 100-500 → API: maara
}

export interface FilterOptions {
  occupation_groups: Array<{
    code: string;
    name_fi: string;
    name_en?: string;
    name_sv?: string;
  }>;
  employer_types: Array<{
    code: string;
    name: string;
    value: string;
  }>;
  working_hours: string[];
  durations: string[];
  regions: string[];
  languages: string[];
}
```

## Configuration

### package.json

```json
{
  "name": "tyomarkkinatori-mcp-server",
  "version": "1.0.0",
  "description": "MCP server for Finnish job market search via Työmarkkinatori.fi",
  "type": "module",
  "main": "dist/index.js",
  "bin": {
    "tyomarkkinatori-mcp": "./dist/index.js"
  },
  "scripts": {
    "build": "tsc",
    "dev": "tsx src/index.ts",
    "start": "node dist/index.js",
    "test": "vitest",
    "lint": "eslint src/",
    "format": "prettier --write src/"
  },
  "keywords": ["mcp", "job-search", "finland", "työmarkkinatori"],
  "dependencies": {
    "@modelcontextprotocol/sdk": "^0.5.0",
    "axios": "^1.6.0",
    "dotenv": "^16.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0",
    "tsx": "^4.0.0",
    "vitest": "^1.0.0",
    "eslint": "^8.0.0",
    "prettier": "^3.0.0"
  }
}
```

### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "lib": ["ES2022"],
    "moduleResolution": "node",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

## Testing Strategy

### Unit Tests
- Test individual functions (validators, parsers, formatters)
- Mock external dependencies
- Test error handling

### Integration Tests
- Test tool handlers with mocked API responses
- Test cache behavior
- Test rate limiter

### E2E Tests
- Test against QA API environment
- Validate OAuth authentication flow
- Test complete workflows
- Verify API response parsing

## Deployment

### Development
```bash
npm install
npm run dev
```

### Production
```bash
npm run build
npm start
```

### Environment Variables
Create a `.env` file:
```bash
# OAuth 2.0 Credentials (from KEHA Centre)
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret
TENANT_ID=your-tenant-id

# API Configuration
API_BASE_URL=https://integraatiot.tyomarkkinatori.fi
# For testing: https://integraatiot-qa.tyomarkkinatori.fi

# Optional settings
CACHE_TTL_MS=900000
RATE_LIMIT_RPS=2
RATE_LIMIT_BURST=10
```

### MCP Client Configuration
Add to Claude Desktop config:
```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "node",
      "args": ["/path/to/mcp/dist/index.js"],
      "env": {
        "CLIENT_ID": "your-client-id",
        "CLIENT_SECRET": "your-client-secret",
        "TENANT_ID": "your-tenant-id"
      }
    }
  }
}
```

## Maintenance

### API Changes
- Monitor Työmarkkinatori API documentation for updates
- Test with QA environment before deploying changes
- Subscribe to API changelog if available
- Handle API versioning properly

### Credential Management
- Rotate credentials periodically
- Monitor token expiration and refresh
- Secure storage of client secrets
- Never commit credentials to repository

### Performance Monitoring
- Log response times
- Track cache hit rates
- Monitor error rates
- Track OAuth token refresh frequency
- Monitor API quota usage

### Updates
- Keep dependencies updated (especially @modelcontextprotocol/sdk)
- Test after API updates
- Document breaking changes
- Maintain changelog

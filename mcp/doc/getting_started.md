# Getting Started - Työmarkkinatori MCP Server

This guide will help you implement the Työmarkkinatori MCP server from scratch.

## Prerequisites

- Node.js 18+ or Bun
- npm or yarn
- Claude Desktop (for testing)
- Basic understanding of TypeScript
- Familiarity with web scraping concepts

## Step-by-Step Implementation

### Phase 1: Project Setup (30 minutes)

#### 1. Initialize Project

```bash
cd mcp
npm init -y
```

#### 2. Install Dependencies

```bash
# Core dependencies
npm install @modelcontextprotocol/sdk axios cheerio

# Development dependencies
npm install -D typescript @types/node tsx vitest eslint prettier
```

#### 3. Configure TypeScript

Create `tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "moduleResolution": "node",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true
  }
}
```

#### 4. Update package.json

```json
{
  "name": "tyomarkkinatori-mcp-server",
  "version": "1.0.0",
  "type": "module",
  "main": "dist/index.js",
  "bin": {
    "tyomarkkinatori-mcp": "./dist/index.js"
  },
  "scripts": {
    "build": "tsc",
    "dev": "tsx src/index.ts",
    "start": "node dist/index.js"
  }
}
```

### Phase 2: Core Implementation (2-3 hours)

#### 1. Create Directory Structure

```bash
mkdir -p src/{tools,client,models,utils}
```

#### 2. Implement Data Models (src/models/job.ts)

Start with the basic interfaces:
```typescript
export interface JobListing {
  id: string;
  title: string;
  employer: string;
  location: string;
  employment_type: string;
  published_date: string;
  summary: string;
  url: string;
}

export interface SearchParams {
  query?: string;
  location?: string;
  category?: string;
  employment_type?: string;
  published_after?: string;
  limit?: number;
}
```

#### 3. Implement Utilities

**Cache (src/utils/cache.ts):**
- Implement basic Map-based cache
- Add TTL support
- Add cleanup mechanism

**Rate Limiter (src/utils/rate_limiter.ts):**
- Token bucket algorithm
- Async waiting mechanism

**Logger (src/utils/logger.ts):**
- Simple console logging
- Different log levels

#### 4. Implement Web Scraper (src/client/scraper.ts)

Key steps:
1. Fetch the actual job search page
2. Inspect HTML structure using browser DevTools
3. Identify CSS selectors for:
   - Job listing cards
   - Job title
   - Employer name
   - Location
   - Employment type
   - Publication date
   - Job URL
4. Implement parsing logic with cheerio

**Testing scraper separately:**
```bash
# Create a test script
node -e "
import { JobScraper } from './src/client/scraper.js';
import axios from 'axios';

const response = await axios.get('https://tyomarkkinatori.fi/avoimet-tyopaikat');
const scraper = new JobScraper();
const jobs = scraper.parseSearchResults(response.data);
console.log(jobs);
"
```

#### 5. Implement API Client (src/client/api_client.ts)

- Wrap axios with rate limiting
- Integrate cache
- Connect scraper
- Handle errors gracefully

### Phase 3: MCP Server (1-2 hours)

#### 1. Implement Server (src/server.ts)

Follow the template in [implementation.md](implementation.md):
- Set up MCP Server instance
- Register tool handlers
- Implement request routing

#### 2. Implement Tool Handlers (src/tools/)

**search_jobs.ts:**
- Validate parameters
- Call API client
- Format response

**get_job_details.ts:**
- Validate job_id
- Call API client
- Return detailed info

**list_categories.ts:**
- Return hardcoded or scraped categories
- Can be enhanced later with dynamic scraping

#### 3. Create Entry Point (src/index.ts)

- Set up stdio transport
- Handle errors
- Add logging

### Phase 4: Testing & Debugging (1-2 hours)

#### 1. Manual Testing

```bash
npm run dev
```

Test with echo:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npm run dev
```

#### 2. Test with Claude Desktop

Add to config:
```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "node",
      "args": ["/absolute/path/to/mcp/dist/index.js"]
    }
  }
}
```

Restart Claude Desktop and test:
- "Search for developer jobs in Helsinki"
- "Find nursing jobs posted last week"

#### 3. Debug Common Issues

**HTML Parsing Issues:**
- Selectors don't match → Update selectors
- Missing data → Add null checks
- Wrong encoding → Set charset

**MCP Protocol Issues:**
- Wrong response format → Check MCP SDK docs
- Tools not appearing → Verify tool registration
- Connection issues → Check stdio setup

### Phase 5: Enhancement (Ongoing)

#### Short-term
1. Improve error messages
2. Add more filters
3. Better date parsing
4. Handle pagination

#### Medium-term
1. Add unit tests
2. Improve caching strategy
3. Add configuration file
4. Create detailed logging

#### Long-term
1. Job alerts
2. Application tracking
3. Multiple job boards
4. Analytics

## Testing Checklist

- [ ] Server starts without errors
- [ ] Tools list returns all 3 tools
- [ ] Basic job search works
- [ ] Search with filters works
- [ ] Job details retrieval works
- [ ] Categories listing works
- [ ] Rate limiting is working
- [ ] Caching is working
- [ ] Error handling is graceful
- [ ] Works with Claude Desktop

## Common Issues & Solutions

### Issue: Selectors Not Matching

**Solution:**
1. Open https://tyomarkkinatori.fi/avoimet-tyopaikat in browser
2. Right-click → Inspect Element
3. Find job listing elements
4. Update selectors in scraper.ts

### Issue: Website Blocking Requests

**Solution:**
- Add realistic User-Agent
- Implement rate limiting
- Add delays between requests
- Use rotating proxies (advanced)

### Issue: Date Parsing Failures

**Solution:**
- Test different date formats
- Add fallback to current date
- Log unparsed dates for analysis

### Issue: Empty Results

**Solution:**
- Check if website structure changed
- Verify selectors are correct
- Add logging to see what's being parsed
- Test with curl to see raw response

## Development Workflow

### Daily Development
```bash
# 1. Start development server
npm run dev

# 2. Make changes
# 3. Test with echo or Claude Desktop
# 4. Check logs for errors
# 5. Iterate
```

### Before Committing
```bash
npm run lint
npm run format
npm run build
npm test
```

### Deployment
```bash
npm run build
# Test the built version
node dist/index.js
```

## Next Steps

1. **Week 1:** Core implementation (Phases 1-3)
2. **Week 2:** Testing and debugging (Phase 4)
3. **Week 3:** Enhancement and polish (Phase 5)
4. **Week 4:** Documentation and release

## Resources

- [MCP Documentation](https://modelcontextprotocol.io/)
- [Cheerio Documentation](https://cheerio.js.org/)
- [Axios Documentation](https://axios-http.com/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)

## Support

If you get stuck:
1. Check error messages carefully
2. Review documentation
3. Test components individually
4. Add console.log for debugging
5. Verify HTML structure hasn't changed

## Tips for Success

1. **Start Simple:** Get basic search working first
2. **Test Often:** Test after each component
3. **Log Everything:** Add detailed logging
4. **Handle Errors:** Always have fallbacks
5. **Document Changes:** Note when website changes
6. **Respect Website:** Use rate limiting
7. **Cache Aggressively:** Reduce unnecessary requests
8. **Validate Input:** Prevent invalid queries
9. **Monitor Performance:** Track response times
10. **Keep It Maintainable:** Write clean, documented code

## Estimated Timeline

- **Minimal Working Version:** 4-6 hours
- **Polished Version:** 15-20 hours
- **Production Ready:** 30-40 hours

Good luck with your implementation! 🚀

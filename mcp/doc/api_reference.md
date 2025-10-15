# Työmarkkinatori MCP Server - API Reference

## Overview

This document provides a comprehensive reference for all tools exposed by the Työmarkkinatori MCP server.

## Tools

### 1. search_jobs

Search for job listings on Työmarkkinatori.fi based on various criteria using the official API.

#### Parameters

| Parameter | Type | Required | Default | Description | API Mapping |
|-----------|------|----------|---------|-------------|-------------|
| query | string | No | - | Keywords to search for (job title, skills, description). Example: "ohjelmistokehittäjä", "sairaanhoitaja" | `hakusana` |
| location | string | No | - | Location filter. Examples: "Helsinki", "Tampere", "Uusimaa", "Pirkanmaa" | `sijainti` |
| occupation_group | string | No | - | ESCO occupation group code or name. Example: "2512" (Software developers) | `ammattiryhmä` |
| employer_type | string | No | - | Type of employer. Values: "company", "public", "nonprofit" | `tyollistaja` (01/02/03) |
| working_hours | string | No | - | Working hours type. Values: "full-time", "part-time" | `työaika` |
| duration | string | No | - | Employment duration. Values: "permanent", "temporary" | `työn jatkuvuus` |
| published_after | string | No | - | Only show jobs published after this date. Format: ISO 8601 date string (e.g., "2024-01-15") | `julkaisupvm` |
| language | string | No | "fi" | Language for results. Values: "fi", "sv", "en" | `kieli` |
| page | number | No | 0 | Page number for pagination (0-based) | `sivu` |
| page_size | number | No | 100 | Results per page. Range: 100-500 | `maara` |

#### Returns

Returns a JSON object containing:

```typescript
{
  total: number,           // Total number of results found
  jobs: JobListing[]       // Array of job listings
}
```

Where `JobListing` has the structure:

```typescript
{
  id: string,                      // Unique job identifier
  title: string,                   // Job title
  employer: string,                // Company/employer name
  location: string,                // Job location
  employment_type: string,         // Employment type
  published_date: string,          // ISO 8601 date string
  application_deadline?: string,   // ISO 8601 date string (optional)
  summary: string,                 // Brief job description
  url: string                      // Direct link to job posting
}
```

#### Example Usage

**Basic search:**
```json
{
  "query": "ohjelmistokehittäjä"
}
```

**Advanced search with filters:**
```json
{
  "query": "sairaanhoitaja",
  "location": "Helsinki",
  "working_hours": "full-time",
  "duration": "permanent",
  "employer_type": "public",
  "published_after": "2024-01-01",
  "page_size": 100
}
```

**ESCO occupation group search:**
```json
{
  "occupation_group": "2512",
  "location": "Uusimaa",
  "page_size": 200
}
```

**Pagination example:**
```json
{
  "query": "insinööri",
  "page": 2,
  "page_size": 100
}
```

**Multi-language search:**
```json
{
  "query": "developer",
  "language": "en",
  "location": "Helsinki"
}
```

#### Example Response

```json
{
  "total": 2,
  "jobs": [
    {
      "id": "12345",
      "title": "Senior Ohjelmistokehittäjä",
      "employer": "TechFinn Oy",
      "location": "Helsinki",
      "employment_type": "kokoaikainen",
      "published_date": "2024-03-15T08:00:00Z",
      "application_deadline": "2024-04-15T23:59:59Z",
      "summary": "Etsimme kokeneita ohjelmistokehittäjiä meidän tiimiin...",
      "url": "https://tyomarkkinatori.fi/tyopaikka/12345"
    },
    {
      "id": "12346",
      "title": "Full Stack Developer",
      "employer": "Digital Solutions Finland",
      "location": "Helsinki",
      "employment_type": "kokoaikainen",
      "published_date": "2024-03-14T10:30:00Z",
      "summary": "Haemme full stack kehittäjää kasvavaan tiimiimme...",
      "url": "https://tyomarkkinatori.fi/tyopaikka/12346"
    }
  ]
}
```

#### Error Responses

```json
{
  "error": "Failed to search jobs",
  "message": "Network request failed"
}
```

```json
{
  "error": "Invalid parameters",
  "message": "limit must be between 1 and 100"
}
```

---

### 2. get_job_details

Retrieve detailed information about a specific job listing.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| job_id | string | Yes | Unique identifier for the job listing |

#### Returns

Returns a complete `JobListing` object with extended fields:

```typescript
{
  id: string,
  title: string,
  employer: string,
  location: string,
  employment_type: string,
  published_date: string,
  application_deadline?: string,
  summary: string,
  description?: string,           // Full HTML description
  requirements?: string[],        // List of requirements
  benefits?: string[],            // List of benefits
  salary_info?: string,           // Salary information if available
  url: string
}
```

#### Example Usage

```json
{
  "job_id": "12345"
}
```

#### Example Response

```json
{
  "id": "12345",
  "title": "Senior Ohjelmistokehittäjä",
  "employer": "TechFinn Oy",
  "location": "Helsinki",
  "employment_type": "kokoaikainen",
  "published_date": "2024-03-15T08:00:00Z",
  "application_deadline": "2024-04-15T23:59:59Z",
  "summary": "Etsimme kokeneita ohjelmistokehittäjiä meidän tiimiin...",
  "description": "<p>TechFinn Oy on kasvava teknologiayritys...</p><p>Tehtävässä vastaat...</p>",
  "requirements": [
    "5+ vuotta kokemusta ohjelmistokehityksestä",
    "Vahva osaaminen TypeScript/Node.js",
    "Kokemus React/Vue kehityksestä",
    "Hyvät suomen kielen taidot"
  ],
  "benefits": [
    "Kilpailukykyinen palkka",
    "Etätyömahdollisuus",
    "Kattava koulutuspaketti",
    "Henkilöstöedut"
  ],
  "salary_info": "4500-6500 EUR/kk",
  "url": "https://tyomarkkinatori.fi/tyopaikka/12345"
}
```

#### Error Responses

```json
{
  "error": "Job not found",
  "message": "No job listing found with ID: 12345"
}
```

---

### 3. list_categories

Get available filter options including occupation groups (ESCO), employer types, working hours, durations, regions, and languages.

#### Parameters

None

#### Returns

```typescript
{
  occupation_groups: Array<{     // ESCO occupation groups
    code: string,                // ESCO code (e.g., "2512")
    name_fi: string,             // Finnish name
    name_en?: string,            // English name
    name_sv?: string             // Swedish name
  }>,
  employer_types: Array<{        // Types of employers
    code: string,                // API code (01, 02, 03)
    name: string,                // Display name
    value: string                // MCP parameter value
  }>,
  working_hours: string[],       // ["full-time", "part-time"]
  durations: string[],           // ["permanent", "temporary"]
  regions: string[],             // Finnish regions (maakunnat)
  languages: string[]            // ["fi", "sv", "en"]
}
```

#### Example Usage

```json
{}
```

#### Example Response

```json
{
  "occupation_groups": [
    {
      "code": "2512",
      "name_fi": "Ohjelmistokehittäjät",
      "name_en": "Software developers"
    },
    {
      "code": "2221",
      "name_fi": "Sairaanhoitajat",
      "name_en": "Nursing professionals"
    }
  ],
  "employer_types": [
    {
      "code": "01",
      "name": "Yritykset",
      "value": "company"
    },
    {
      "code": "02",
      "name": "Julkinen sektori",
      "value": "public"
    },
    {
      "code": "03",
      "name": "Järjestöt",
      "value": "nonprofit"
    }
  ],
  "working_hours": [
    "full-time",
    "part-time"
  ],
  "durations": [
    "permanent",
    "temporary"
  ],
  "regions": [
    "Uusimaa",
    "Pirkanmaa",
    "Varsinais-Suomi",
    "Pohjois-Pohjanmaa",
    "Keski-Suomi",
    "Pohjois-Savo",
    "Kanta-Häme",
    "Satakunta",
    "Päijät-Häme",
    "Lappi"
  ],
  "languages": [
    "fi",
    "sv",
    "en"
  ]
}
```

---

## Data Models

### JobListing

Complete structure for a job listing:

```typescript
interface JobListing {
  // Basic information
  id: string;                        // Unique identifier
  title: string;                     // Job title
  employer: string;                  // Employer name
  location: string;                  // Job location
  employment_type: string;           // Employment type

  // Dates
  published_date: string;            // ISO 8601 format
  application_deadline?: string;     // ISO 8601 format (optional)

  // Description
  summary: string;                   // Brief summary
  description?: string;              // Full description (HTML, in details only)

  // Details (available in get_job_details)
  requirements?: string[];           // List of job requirements
  benefits?: string[];               // List of benefits offered
  salary_info?: string;              // Salary information

  // Link
  url: string;                       // Direct URL to job posting
}
```

### SearchParams

Parameters for job search:

```typescript
interface SearchParams {
  // Search & Basic Filters
  query?: string;              // Keyword search (maps to API: hakusana)
  location?: string;           // Location filter (maps to API: sijainti)
  occupation_group?: string;   // ESCO occupation group (maps to API: ammattiryhmä)

  // Employment Filters
  employer_type?: "company" | "public" | "nonprofit";  // Employer type (maps to API: tyollistaja)
  working_hours?: "full-time" | "part-time";           // Working hours (maps to API: työaika)
  duration?: "permanent" | "temporary";                // Duration (maps to API: työn jatkuvuus)

  // Date & Language
  published_after?: string;    // Date filter ISO 8601 (maps to API: julkaisupvm)
  language?: "fi" | "sv" | "en";  // Language (maps to API: kieli)

  // Pagination
  page?: number;               // Page number 0-based (maps to API: sivu)
  page_size?: number;          // Results per page 100-500 (maps to API: maara)
}
```

---

## Common Patterns

### Finding Recent Jobs in a Specific Field

```json
{
  "query": "sairaanhoitaja",
  "published_after": "2024-03-01",
  "page_size": 200
}
```

### Finding Jobs in Multiple Cities

Make separate calls for each city or use broader region filter:

```json
{
  "query": "myyjä",
  "location": "Uusimaa"
}
```

### Getting Full Details for Interesting Jobs

1. First search for jobs:
```json
{
  "query": "ohjelmistokehittäjä",
  "location": "Helsinki"
}
```

2. Then get details for specific jobs:
```json
{
  "job_id": "12345"
}
```

### Exploring Available Options

Get available filters before searching:
```json
{}  // Call list_categories
```

Then use returned values in search.

---

## Rate Limiting

The server implements rate limiting to respect the API:

- **Default rate:** 2 requests per second
- **Burst capacity:** 10 requests
- **Caching:** Results cached for 15 minutes (searches) and 1 hour (details)

Best practices:
- Use caching effectively (identical searches return cached results)
- Use pagination instead of large page_size values
- Batch your searches logically
- Use appropriate `page_size` values (100-500) to balance performance and API load

---

## Error Handling

### Common Error Types

1. **Authentication Errors**
   - Invalid OAuth credentials
   - Token expired
   - Missing credentials
   - Insufficient permissions

2. **Validation Errors**
   - Invalid parameter format
   - Out-of-range values (e.g., page_size not 100-500)
   - Missing required parameters

3. **Network Errors**
   - Connection timeout
   - DNS resolution failure
   - SSL certificate errors

4. **API Errors**
   - Invalid API response
   - API endpoint changed
   - Unexpected response format
   - Missing expected data

5. **Rate Limit Errors**
   - Too many requests in short time
   - API quota exceeded
   - Server temporarily unavailable

### Error Response Format

```typescript
{
  error: string,      // Error type/category
  message: string,    // Human-readable error message
  details?: any       // Additional error details (optional)
}
```

---

## Caching Behavior

### Cache Keys
- Search results: Based on all search parameters
- Job details: Based on job ID
- Categories: Single cache entry

### Cache TTL (Time To Live)
- Search results: 15 minutes
- Job details: 1 hour
- Categories: 24 hours

### Cache Invalidation
- Automatic expiration based on TTL
- Background cleanup every minute
- Manual invalidation on errors (optional)

---

## API Parameter Mapping

This section documents how MCP tool parameters map to the actual Työmarkkinatori API parameters.

### search_jobs Parameter Mapping

| MCP Parameter | API Parameter | Value Mapping | Example |
|---------------|---------------|---------------|---------|
| query | hakusana | Direct string | "ohjelmistokehittäjä" → hakusana=ohjelmistokehittäjä |
| location | sijainti | Direct string | "Helsinki" → sijainti=Helsinki |
| occupation_group | ammattiryhmä | ESCO code or name | "2512" → ammattiryhmä=2512 |
| employer_type | tyollistaja | company→01, public→02, nonprofit→03 | "company" → tyollistaja=01 |
| working_hours | työaika | full-time→01, part-time→02 | "full-time" → työaika=01 |
| duration | työn jatkuvuus | permanent→01, temporary→02 | "permanent" → työn jatkuvuus=01 |
| published_after | julkaisupvm | ISO 8601 date | "2024-01-15" → julkaisupvm=2024-01-15 |
| language | kieli | Direct: fi, sv, en | "fi" → kieli=fi |
| page | sivu | Direct number (0-based) | 2 → sivu=2 |
| page_size | maara | Direct number (100-500) | 200 → maara=200 |

### Example API Request

**MCP Tool Call:**
```json
{
  "query": "developer",
  "location": "Helsinki",
  "employer_type": "company",
  "working_hours": "full-time",
  "language": "en",
  "page": 0,
  "page_size": 100
}
```

**Actual API Request:**
```
GET /jobpostingprovider/v1/tyopaikat?
  hakusana=developer&
  sijainti=Helsinki&
  tyollistaja=01&
  työaika=01&
  kieli=en&
  sivu=0&
  maara=100
```

### Code Mappings

The implementation uses these mapping tables:

```typescript
// Employer type mapping
const employerTypeMap = {
  "company": "01",
  "public": "02",
  "nonprofit": "03"
};

// Working hours mapping
const workingHoursMap = {
  "full-time": "01",
  "part-time": "02"
};

// Duration mapping
const durationMap = {
  "permanent": "01",
  "temporary": "02"
};
```

---

## Finnish Job Market Terminology

### Employment Types
- **kokoaikainen** - Full-time
- **osa-aikainen** - Part-time
- **määräaikainen** - Fixed-term/Temporary
- **vuokratyö** - Contract work/Temp agency
- **harjoittelu** - Internship/Traineeship

### Common Search Terms
- **kehittäjä** - Developer
- **insinööri** - Engineer
- **hoitaja** - Nurse/Caregiver
- **opettaja** - Teacher
- **myyjä** - Salesperson
- **sihteeri** - Secretary
- **varastotyöntekijä** - Warehouse worker
- **ravintolatyöntekijä** - Restaurant worker

### Regions (Maakunta)
- **Uusimaa** - Helsinki metropolitan area
- **Pirkanmaa** - Tampere region
- **Varsinais-Suomi** - Turku region
- **Pohjois-Pohjanmaa** - Oulu region

---

## Version History

### Version 1.0.0
- Initial release
- Basic search functionality
- Job details retrieval
- Category listing
- Rate limiting and caching

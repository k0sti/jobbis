# Työmarkkinatori API Credentials - Setup Guide

This guide explains how to obtain and configure API credentials for the Työmarkkinatori MCP Server.

## Overview

The Työmarkkinatori API uses **OAuth 2.0 Client Credentials Flow** via Microsoft Identity Platform for authentication. You must obtain credentials from KEHA Centre before using this server.

## Step 1: Request API Access

### Contact Information
- **Email:** tmt-rajapinnat.keha@ely-keskus.fi
- **Subject:** "Työmarkkinatori API Access Request"

### Information to Provide

In your email, include:

1. **Organization Details:**
   - Organization name
   - Business ID (Y-tunnus)
   - Contact person name
   - Contact email and phone

2. **Use Case Description:**
   - Purpose of API usage
   - Expected usage volume
   - Technical implementation overview
   - Data handling practices

3. **Agreement:**
   - Confirm acceptance of Työmarkkinatori API terms of use
   - Confirm compliance with data protection regulations (GDPR)

### Example Email Template

```
Subject: Työmarkkinatori API Access Request

Hyvä KEHA-keskus,

Haluamme pyytää pääsyä Työmarkkinatori-rajapintaan seuraavilla tiedoilla:

Organisaatio: [Your Organization Name]
Y-tunnus: [Business ID]
Yhteyshenkilö: [Your Name]
Sähköposti: [Your Email]
Puhelin: [Your Phone]

Käyttötarkoitus:
[Describe your use case - e.g., "AI-assistentti työnhakijoille"]

Odotettu käyttömäärä:
[e.g., "Noin 100-500 hakua päivässä"]

Tekninen toteutus:
[Brief description - e.g., "MCP-palvelin Claude AI:lle"]

Vakuutamme noudattavamme Työmarkkinatorin API:n käyttöehtoja ja
tietosuojalainsäädäntöä (GDPR).

Ystävällisin terveisin,
[Your Name]
```

## Step 2: Receive QA Environment Credentials

KEHA Centre will respond with **test environment credentials**:

```
CLIENT_ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
CLIENT_SECRET: xxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TENANT_ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
QA_API_URL: https://integraatiot-qa.tyomarkkinatori.fi
```

**Important:** Test thoroughly in the QA environment before requesting production access.

## Step 3: Configure QA Environment

### Configure Claude Desktop

Add to Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "/absolute/path/to/mcp/bin/tyomarkkinatori-mcp",
      "env": {
        "CLIENT_ID": "your-qa-client-id",
        "CLIENT_SECRET": "your-qa-client-secret",
        "TENANT_ID": "your-qa-tenant-id",
        "API_BASE_URL": "https://integraatiot-qa.tyomarkkinatori.fi"
      }
    }
  }
}
```

### Build and Test

```bash
just build
just run
```

If successful, you should see:
```
Työmarkkinatori MCP server running on stdio
OAuth token obtained successfully
```

## Step 4: Test in QA Environment

### Test Checklist

- [ ] Server starts without errors
- [ ] OAuth authentication succeeds
- [ ] search_jobs returns results
- [ ] get_job_details works
- [ ] list_categories returns data
- [ ] Rate limiting functions correctly
- [ ] Error handling works
- [ ] Cache reduces API calls

### Testing with Claude Desktop

1. Configure Claude Desktop with QA credentials
2. Test queries:
   - "Search for IT jobs in Helsinki"
   - "Find nursing jobs posted this week"
   - "Show me available job categories"

3. Monitor logs for errors

### API Testing

Example request (using curl):
```bash
# Get OAuth token
TOKEN=$(curl -X POST "https://login.microsoftonline.com/$TENANT_ID/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "scope=https://graph.microsoft.com/.default" \
  -d "grant_type=client_credentials" \
  | jq -r '.access_token')

# Test API
curl "https://integraatiot-qa.tyomarkkinatori.fi/jobpostingprovider/v1/tyopaikat?sivu=0&maara=10&kieli=fi" \
  -H "Authorization: Bearer $TOKEN"
```

## Step 5: Request Production Access

### After Successful QA Testing

Email KEHA Centre again:
```
Subject: Työmarkkinatori API Production Access Request

Hyvä KEHA-keskus,

Olemme testanneet API-integraatiomme onnistuneesti QA-ympäristössä
ja haluamme pyytää tuotantoympäristön tunnukset.

Testit suoritettu:
✓ OAuth-autentikointi
✓ Työpaikkojen haku
✓ Työpaikan tietojen haku
✓ Kategorioiden listaus
✓ Virheenkäsittely

Organisaatio: [Your Organization]
QA Client ID: [Your QA Client ID]

Ystävällisin terveisin,
[Your Name]
```

## Step 6: Configure Production Environment

### Receive Production Credentials

You'll receive:
```
PROD_CLIENT_ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
PROD_CLIENT_SECRET: xxxxxxxxxxxxxxxxxxxxxxxxxxxxx
PROD_TENANT_ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
PROD_API_URL: https://integraatiot.tyomarkkinatori.fi
```

### Update Claude Desktop Configuration

Update `claude_desktop_config.json` with production credentials:

```json
{
  "mcpServers": {
    "tyomarkkinatori": {
      "command": "/absolute/path/to/mcp/bin/tyomarkkinatori-mcp",
      "env": {
        "CLIENT_ID": "your-prod-client-id",
        "CLIENT_SECRET": "your-prod-client-secret",
        "TENANT_ID": "your-prod-tenant-id",
        "API_BASE_URL": "https://integraatiot.tyomarkkinatori.fi"
      }
    }
  }
}
```

### Start Using Production

Restart Claude Desktop to load the new configuration.

## Security Best Practices

### Credential Storage

✅ **DO:**
- Store credentials in MCP client configuration
- Use environment variables via MCP client config
- Rotate credentials periodically
- Keep MCP config file secure

❌ **DON'T:**
- Commit credentials to git
- Share credentials publicly
- Hard-code credentials
- Log full credentials

### Access Control

- Limit who has access to credentials
- Use separate credentials for different environments
- Monitor API usage for anomalies
- Set up alerts for authentication failures

### Credential Rotation

Rotate credentials:
- Every 90 days (recommended)
- When team members with access leave
- If credentials are potentially compromised
- As required by your security policy

## Troubleshooting

### Authentication Fails

**Problem:** "Authentication failed - check credentials"

**Solutions:**
1. Verify credentials are correct
2. Check if credentials are for correct environment (QA vs Production)
3. Ensure no extra spaces in MCP config
4. Verify `TENANT_ID` is correct

**Test credentials separately:**
```bash
curl -X POST "https://login.microsoftonline.com/$TENANT_ID/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "scope=https://graph.microsoft.com/.default" \
  -d "grant_type=client_credentials"
```

### Token Expires Too Quickly

**Problem:** Frequent token refresh requests

**Solution:**
- This is normal (tokens expire after ~1 hour)
- The server automatically refreshes tokens
- Check logs for refresh frequency
- Ensure token caching is working

### API Returns 401 Unauthorized

**Problem:** API requests fail with 401

**Causes:**
1. Token expired (should auto-refresh)
2. Wrong API endpoint
3. Credentials revoked
4. Token for wrong tenant

**Solution:**
1. Check token expiry logic in code
2. Verify API_BASE_URL is correct
3. Contact KEHA Centre to verify credentials status

### Rate Limiting Issues

**Problem:** Too many requests error

**Solution:**
1. Increase `RATE_LIMIT_RPS` in MCP config
2. Check if cache is working properly
3. Review API usage patterns
4. Contact KEHA Centre about quota increase

## Support

### Technical Issues

For technical problems with the MCP server:
- Check logs in server output
- Review documentation in `doc/` folder
- Test with simpler queries

### API Access Issues

For credential or API access issues:
- **Email:** tmt-rajapinnat.keha@ely-keskus.fi
- **Subject:** Include your organization name and client ID

### Common Questions

**Q: How long do credentials remain valid?**
A: Credentials don't expire automatically, but should be rotated periodically (every 90 days recommended).

**Q: Can I use the same credentials for multiple servers?**
A: Yes, but monitor usage to stay within quota limits.

**Q: What's the API rate limit?**
A: Rate limits are not explicitly published. Be respectful with request frequency and use caching.

**Q: Can I get a sandbox/testing account?**
A: Yes, QA environment serves as the testing environment.

**Q: Do I need separate credentials for each developer?**
A: No, but consider separate credentials for development vs production environments.

## Checklist

Before going to production:

- [ ] Received QA credentials from KEHA Centre
- [ ] Tested all functionality in QA environment
- [ ] Implemented proper error handling
- [ ] Enabled caching
- [ ] Configured rate limiting
- [ ] Secured credential storage
- [ ] Set up monitoring and logging
- [ ] Received production credentials
- [ ] Tested in production with limited scope
- [ ] Documented deployment process
- [ ] Set up credential rotation schedule

## Next Steps

Once credentials are configured:

1. Follow [Implementation Guide (Go)](implementation-go.md) for development
2. Read [API Reference](api_reference.md) for tool usage
3. Check [Architecture Document](architecture.md) for system design

---

*Last Updated: 2025-10-15*
*For the latest API documentation, visit Työmarkkinatori or contact KEHA Centre*

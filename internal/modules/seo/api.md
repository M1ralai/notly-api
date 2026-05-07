# SEO Module API

## Overview
The SEO module provides programmatic SEO capabilities through a database-driven approach. All SEO pages are stored in PostgreSQL and dynamically served via API endpoints.

## Endpoints

### GET /api/seo/pages
Fetch all active SEO pages.

**Response:**
```json
[
  {
    "id": 1,
    "slug": "yks-puan-hesaplama",
    "type": "tool",
    "title": "YKS Puan Hesaplama 2026",
    "meta_description": "TYT ve AYT netlerinizi girin...",
    "meta_keywords": ["yks puan hesaplama", "tyt ayt"],
    "content_config": {"tool_type": "yks_calculator"},
    "priority": 0.95,
    "changefreq": "monthly",
    "is_active": true,
    "created_at": "2026-01-10T00:00:00Z",
    "updated_at": "2026-01-10T00:00:00Z"
  }
]
```

### GET /api/seo/pages/{slug}
Fetch a single SEO page by slug.

**Parameters:**
- `slug` (path) - URL-safe slug (e.g., `yks-puan-hesaplama`)

**Response:**
```json
{
  "id": 1,
  "slug": "yks-puan-hesaplama",
  "type": "tool",
  ...
}
```

**Status Codes:**
- `200` - Page found
- `404` - Page not found
- `500` - Server error

### GET /api/seo/pages/{id}/related
Fetch related pages for a given page ID (Internal Linking).

**Parameters:**
- `id` (path) - Page ID
- `limit` (query) - Max number of results (default: 6)

**Response:**
```json
[
  {
    "id": 2,
    "slug": "tyt-puan-hesaplama",
    "title": "TYT Puan Hesaplama 2026",
     ...
  }
]
```

### GET /embed/{slug}
Serve an embeddable calculator widget with backlinks.

**Parameters:**
- `slug` (path) - Tool slug

**Response:**
Returns HTML content suitable for iframe embedding.

### GET /sitemap.xml
Generate dynamic XML sitemap from database.

**Response:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://notly.app/</loc>
    <lastmod>2026-01-10</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.00</priority>
  </url>
  ...
</urlset>
```

**Cache Headers:**
- `Cache-Control: public, max-age=3600, stale-while-revalidate=86400`

## Page Types
- `core` - Core application pages (/, /tasks, /habits, etc.)
- `tool` - Calculator tools (/araclar/:slug)
- `template` - Study templates (/sablonlar/:slug)
- `guide` - Educational guides (/rehber/:slug)
- `feature` - Feature pages (/ozellikler/:slug)

# Public deployment checklist

## 1. GitHub

Repository: `https://github.com/Aravindh97/aravindh-impact-portfolio`

The source is already configured with the Go module:

```text
github.com/Aravindh97/aravindh-impact-portfolio
```

Push `main` only after reviewing the first local commit. GitHub Actions runs Go, Python and Next.js validation.

## 2. Render APIs

Use `render.yaml` as a Blueprint, or create its two services manually. Record the resulting URLs:

```text
https://aravindh-humanity-core.onrender.com
https://aravindh-humanity-intelligence.onrender.com
```

Verify both `/healthz` endpoints before deploying the frontend.

## 3. Vercel frontend

Import the GitHub repository and set the root directory to `apps/web`. Add production, preview and development environment variables:

```text
CORE_API_URL=<deployed Go URL>
INTELLIGENCE_API_URL=<deployed Python URL>
```

Deploy, then verify all six `/systems/<slug>` routes and `/resume.pdf` in a signed-out/incognito browser.

## 4. Final resume URL

Add the Vercel URL as `site_url` in `devresume/resume.json`, run `make resume`, commit the JSON and PDF, and push again.

## Safety

All current endpoints use public, sourced or simulated data. Do not add child, health, animal-owner, radiation-monitoring or precise wildlife data before implementing reviewed identity, authorization, encryption, retention and audit controls.

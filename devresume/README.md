# Resume PDF generator

A small Go command that validates `resume.json` and produces the PDF served by the canonical Next.js frontend.

```bash
go run ./devresume/cmd/resume \
  -data devresume/resume.json \
  -out apps/web/public/resume.pdf
```

The website itself is in `apps/web`. This package has no browser runtime and no network listener.

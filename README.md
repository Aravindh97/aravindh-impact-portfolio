# Six Impact Platforms — Aravindh Varadharaj

Exactly three Python and three Go projects, each with a distinct Next.js/React interface and tested API.

| Project | Backend | Working interaction |
|---|---|---|
| OpenSchool Enterprise | Python/FastAPI | Generate a governed mastery-aware weekly plan for a learning organization |
| FamineLens | Python/FastAPI | Compare sourced 1876–1878 and 1943 famine history; change modern drivers and uncertainty |
| Kindred | Python/FastAPI | Inspect a privacy-aware animal welfare network and rebuild a human-review queue |
| TerraZero | Go | Inspect live public USGS earthquakes and generate earthquake/volcano/tsunami/flood/fire impact zones |
| Nuclear Legacy & Civil Safety | Go | Compare civilian incidents and explore monitoring/recovery scenarios |
| EcoSignal | Go | Preserve network coverage while reducing configurable transmit power, idle airtime and relative habitat activity |

## Stack

- Next.js 16.4, React 19.3, TypeScript 7 and Tailwind CSS 4
- Python 3.12, FastAPI 0.143 and Pydantic 2.14
- Go 1.25 `net/http`
- Production paths for PostgreSQL/PostGIS, time-series data, object storage, queues and OpenTelemetry

## Safety and evidence boundaries

- FamineLens keeps historical sources, uncertainty and contested interpretations; similarity is not causation.
- Kindred coarsens wildlife locations and never diagnoses animals or automatically accuses people.
- TerraZero scenarios are not predictions or authorized evacuation orders.
- Nuclear Legacy is educational civilian safety only—not a dose calculator, medical instruction, weapon design or targeting system.
- EcoSignal reports a relative RF activity proxy, not a biological dose or wildlife safety certification; calibrated instruments and expert review are required.
- OpenSchool requires open-licensed, age-reviewed content and parent/educator safeguarding.

## Run

```bash
python3 -m venv python/.venv
python/.venv/bin/pip install -r python/requirements.txt
cd apps/web && npm ci && cd ../..
make test PY=python/.venv/bin/python
make resume
```

Use three terminals:

```bash
make dev-core                                      # 127.0.0.1:8090
make dev-intelligence PY=python/.venv/bin/python  # 127.0.0.1:8091
make dev-web                                       # 127.0.0.1:3000
```

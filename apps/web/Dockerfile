# syntax=docker/dockerfile:1
FROM node:24.20.0-alpine AS deps
WORKDIR /app
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci --ignore-scripts --no-audit --no-fund

FROM deps AS build
COPY apps/web/ ./
# Rewrites stay on the private service network.
ENV CORE_API_URL=http://core:8090 INTELLIGENCE_API_URL=http://intelligence:8091 NEXT_TELEMETRY_DISABLED=1
RUN npm run build

FROM node:24.20.0-alpine
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 CORE_API_URL=http://core:8090 INTELLIGENCE_API_URL=http://intelligence:8091
WORKDIR /app
COPY --from=build --chown=node:node /app ./
USER node
EXPOSE 3000
# Container bind is published to host loopback only by compose.yaml.
CMD ["npm", "run", "start", "--", "--hostname", "0.0.0.0", "-p", "3000"]

# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal/ internal/
COPY services/core/ services/core/
RUN CGO_ENABLED=0 go test ./services/core/... \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/humanity-core ./services/core/cmd/humanity-core

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/humanity-core /humanity-core
# Container bind is reachable only on the private Compose/service network; no host port is published.
ENV ADDR=:8090
EXPOSE 8090
USER nonroot:nonroot
ENTRYPOINT ["/humanity-core"]

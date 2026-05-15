FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.work ./
COPY apps/api/go.mod apps/api/go.mod
COPY apps/judge-worker/go.mod apps/judge-worker/go.mod
WORKDIR /src/apps/judge-worker
RUN go mod download
WORKDIR /src
COPY apps/judge-worker apps/judge-worker
RUN go build -o /out/judge-worker ./apps/judge-worker/cmd/worker

FROM alpine:3.21
RUN adduser -D -u 10001 judge
COPY --from=build /out/judge-worker /usr/local/bin/judge-worker
USER judge
CMD ["judge-worker"]

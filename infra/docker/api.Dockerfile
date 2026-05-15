FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.work ./
COPY apps/api/go.mod apps/api/go.mod
COPY apps/judge-worker/go.mod apps/judge-worker/go.mod
WORKDIR /src/apps/api
RUN go mod download
WORKDIR /src
COPY apps/api apps/api
RUN go build -o /out/api ./apps/api/cmd/api

FROM alpine:3.21
COPY --from=build /out/api /usr/local/bin/api
EXPOSE 8080
CMD ["api"]

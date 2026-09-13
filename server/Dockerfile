FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o /out/syncspace ./cmd
RUN go build -v -o /out/worker ./cmd/worker

FROM alpine:3.21
WORKDIR /app
COPY --from=build /out/syncspace /app/syncspace
COPY --from=build /out/worker /app/worker
EXPOSE 8080
CMD ["/app/syncspace"]

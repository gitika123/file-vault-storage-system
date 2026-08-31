FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/seed ./cmd/seed

FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app && mkdir -p /var/lib/balkanid/blobs && chown -R app:app /var/lib/balkanid
USER app
COPY --from=build /out/api /usr/local/bin/balkanid-api
COPY --from=build /out/seed /usr/local/bin/balkanid-seed
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/balkanid-api"]

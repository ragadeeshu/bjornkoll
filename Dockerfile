FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY main.go ./
# timetzdata embeds the timezone database, so setting TZ still works in the
# scratch image below.
RUN CGO_ENABLED=0 go build -trimpath -tags timetzdata -o /bjornkoll .

# Runtime image: only the binary, its template and CA certificates (needed to
# scrape thebearden.se over HTTPS). No shell or OS packages to patch.
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bjornkoll /app/bjornkoll
COPY html /app/html
WORKDIR /app
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/app/bjornkoll"]

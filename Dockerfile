FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go mod init svc-trading-api || true
RUN go get github.com/gorilla/mux github.com/prometheus/client_golang github.com/google/uuid
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./src

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/app /app
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app"]

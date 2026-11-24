# Image chạy một lần: goose + psql.
FROM golang:1.26.2-alpine AS goose
ENV CGO_ENABLED=0
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.29.0

FROM postgres:17.10-alpine
COPY --from=goose /go/bin/goose /usr/local/bin/goose
COPY migrate.sh /usr/local/bin/migrate.sh
# Bỏ CRLF nếu file được checkout trên Windows.
RUN sed -i 's/\r$//' /usr/local/bin/migrate.sh \
 && chmod +x /usr/local/bin/migrate.sh
ENTRYPOINT ["/usr/local/bin/migrate.sh"]

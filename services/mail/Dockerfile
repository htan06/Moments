# Image của mail service. Build từ thư mục services/mail/ (services/compose.yml).
FROM golang:1.26.2-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/mail ./cmd/mail

# distroless/static có sẵn CA cert, cần cho kết nối TLS tới SMTP.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mail /mail
ENTRYPOINT ["/mail"]

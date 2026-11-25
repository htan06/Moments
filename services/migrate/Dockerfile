# Image của lệnh migrate (Go). Build từ thư mục services/migrate/ (services/compose.yml).
# Migration SQL và grants.sql không nằm trong image: compose mount application/internal
# vào /migrations, nên sửa migration / grant không cần build lại image.
FROM golang:1.26.2-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/migrate /migrate
ENTRYPOINT ["/migrate"]

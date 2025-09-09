FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend-server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/backend-server /app/backend-server
ENV PORT=6000
ENV KUBECONFIG=""
EXPOSE 6000
USER 65532:65532
ENTRYPOINT ["/app/backend-server"]




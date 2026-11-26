FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -ldflags="-s -w" -o /manager ./cmd
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /manager /manager
USER 10001:10001
ENTRYPOINT ["/manager"]

FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/archon-api ./cmd/archon-api

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/archon-api /archon-api
EXPOSE 8080
ENTRYPOINT ["/archon-api"]

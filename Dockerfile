FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY tools/go.mod tools/go.sum ./tools/
RUN go mod download && go mod download -modfile=tools/go.mod
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server
# The migrate image pins goose to the version in tools/go.mod.
RUN CGO_ENABLED=0 go build -modfile=tools/go.mod -o /out/goose github.com/pressly/goose/v3/cmd/goose

FROM gcr.io/distroless/static-debian12 AS migrate
COPY --from=build /out/goose /goose
COPY --from=build /src/internal/infra/storage/db/migrate/sql /migrations
ENTRYPOINT ["/goose", "-dir", "/migrations"]

FROM gcr.io/distroless/static-debian12 AS backend
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]

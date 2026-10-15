FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go install github.com/uadmin/uadmin/cmd/uadmin@v0.10.2
COPY . .
# uAdmin reads its HTML templates and static files from the working directory;
# "uadmin prepare" copies them out of the module.
RUN uadmin prepare && go build -trimpath -o /out/petshop .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/petshop /app/petshop
COPY --from=build /src/templates /app/templates
COPY --from=build /src/static /app/static
COPY --from=build /src/media /app/media
EXPOSE 8080
ENTRYPOINT ["/app/petshop"]

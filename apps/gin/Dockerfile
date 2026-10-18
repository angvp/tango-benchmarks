FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/petshop .

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/petshop /petshop
EXPOSE 8080
ENTRYPOINT ["/petshop"]

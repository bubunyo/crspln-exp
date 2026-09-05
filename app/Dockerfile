FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod .
COPY main.go .
RUN CGO_ENABLED=0 go build -o /app .

FROM gcr.io/distroless/static
COPY --from=build /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]

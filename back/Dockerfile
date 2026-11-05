FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /contractcheck ./cmd/api

FROM alpine:3.21
COPY --from=build /contractcheck /contractcheck
EXPOSE 8080
CMD ["/contractcheck"]

FROM golang:1.17-alpine as build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o songs .
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

FROM scratch as run
WORKDIR /app
COPY --from=build /build/songs .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD [ "./songs" ]
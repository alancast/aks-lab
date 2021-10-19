FROM golang:1.17-alpine as build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o api .

FROM scratch as run
WORKDIR /app
COPY --from=build /build/api .
CMD [ "./api" ]
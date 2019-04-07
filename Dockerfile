FROM golang:1.10.1 AS build-env

WORKDIR /go/src/app

COPY ./src .

RUN go get -d ./...

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o picfolio .

# Build runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /app

COPY --from=build-env /go/src/app/picfolio .
COPY --from=build-env /go/src/app/www ./www

EXPOSE 80 8080

ENTRYPOINT ./picfolio
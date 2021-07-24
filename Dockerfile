# build
FROM golang:1.14 AS build

WORKDIR /app

COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o /image-cloner

# deploy
FROM alpine:3.14

WORKDIR /

COPY --from=build /image-cloner /image-cloner

EXPOSE 443

CMD ["/image-cloner"]


FROM golang:alpine

RUN apk update && apk add ffmpeg build-base bash

WORKDIR /app

COPY . .

RUN CGO_ENABLED=1 go build -o musicbot main.go

CMD ["./musicbot", "-bottoken=<bot token here>", "-youtubeapikey=<youtube api key here>"]
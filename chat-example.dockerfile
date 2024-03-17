#base go image

FROM golang:1.21-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o chatApp ./

RUN chmod +x ./chatApp

#build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app /app

WORKDIR /app

ENV PASS=supersecret PORT=80

CMD ["/app/chatApp"]
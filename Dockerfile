#
# immotep
#
# Create with: docker build -t immotep:$(node -p "require('./ui/package.json').version") .
# docker build -t localhost:5000/immotep:0.1.0 .
# run with command line:
#   docker run --rm -it -p 8081:8081 immotep:$(node -p "require('./ui/package.json').version")
# 

FROM golang:1.16.5-alpine AS builder

RUN apk --no-cache add git && apk --no-cache add build-base

WORKDIR /src
COPY ./ui/immotep immotep
COPY ./srv .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o immotepsrv .

FROM alpine:3.14  
RUN apk --no-cache add curl && mkdir /app

WORKDIR /app
COPY --from=builder /src/immotepsrv ./
COPY ./srv/imm.db ./

EXPOSE 8080

CMD ["./immotepsrv", "serve" ] 

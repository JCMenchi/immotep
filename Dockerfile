#
# immotep
#
# Create with: docker build -t immotep:$(node -p "require('./ui/package.json').version") .
# docker build -t localhost:5000/immotep:0.1.0 .
# run with command line:
#   docker run --rm -it -p 8081:8081 immotep:$(node -p "require('./ui/package.json').version")
# 

FROM node:16.8.0 as nodebuilder

COPY ./ui /app
RUN cd /app && npm install && npm run build

FROM golang:1.16.5-alpine AS builder

RUN apk --no-cache add git && apk --no-cache add build-base

WORKDIR /src

COPY ./srv .
COPY --from=nodebuilder /app/immotep api/immotep

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o immotepsrv .

FROM alpine:3.14  
RUN apk --no-cache add curl && adduser immotep -D -h /app

USER immotep
WORKDIR /app

COPY --from=builder /src/immotepsrv ./
#COPY ./srv/imm.db ./

EXPOSE 8080

CMD ["./immotepsrv", "serve" ] 

FROM golang:1.14.3-alpine as build
WORKDIR /src
COPY . .
WORKDIR /src/cmd
CMD CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"'  -o /out/hygo . 

#FROM scratch AS bin 
#COPY 
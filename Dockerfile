FROM golang:stretch
WORKDIR /go/src/core_api

# add source code
ADD . /go/src/core_api

EXPOSE 8002
CMD ["./main","-config","./config.json"]
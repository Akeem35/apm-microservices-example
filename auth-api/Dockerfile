FROM golang:alpine3.13
EXPOSE 8081

WORKDIR /go/src/app
RUN apk update && apk add gcc wget alpine-sdk
RUN apk --no-cache add curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . .

# RUN dep ensure -update golang.org/x/sys

# RUN go mod init auth-api
RUN go get github.com/newrelic/go-agent
RUN go mod tidy
RUN go build -o auth-api

CMD /go/src/app/auth-api


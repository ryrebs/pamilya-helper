FROM golang:1.21.4-alpine

ENV PORT=5000

ENV CGO_ENABLED=1

WORKDIR /pamilyahelper

RUN apk update

RUN apk add gcc 

RUN apk add libc-dev 

COPY public /pamilyahelper/public

COPY server /pamilyahelper/server

COPY main.go /pamilyahelper/

COPY go.mod /pamilyahelper/

RUN go mod tidy

RUN go build

# COPY webapp /pamilyahelper/

# RUN /pamilyahelper/webapp initdb

# RUN /pamilyahelper/webapp loadfixture

# ENTRYPOINT [ "/pamilyahelper/webapp", "serve"]

EXPOSE ${PORT}

ENTRYPOINT [ "/bin/sh"]
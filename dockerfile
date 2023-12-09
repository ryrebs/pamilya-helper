FROM golang:1.21.4-alpine

ENV PORT=5000

ENV EXPOSE_IP=0.0.0.0

WORKDIR /pamilyahelper

COPY public /pamilyahelper/public

COPY dist/webapp /pamilyahelper/

COPY fixtures /pamilyahelper/fixtures

RUN ./webapp initdb

RUN ./webapp loadfixtures

ENTRYPOINT [ "./webapp", "serve"]
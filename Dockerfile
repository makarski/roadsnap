FROM golang:alpine3.16

WORKDIR /roadsnap
ADD . .

RUN go install
RUN mkdir snapshots

ENTRYPOINT [ "roadsnap" ]

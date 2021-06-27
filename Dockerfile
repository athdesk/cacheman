FROM golang:1.16-alpine as bld
WORKDIR /go/src/cacheman
COPY . .
RUN go mod download && go get cacheman/local
RUN go build 

FROM alpine
WORKDIR /root
COPY --from=bld /go/src/cacheman/cacheman .
COPY ./default.conf /etc/cacheman/cacheman.conf
EXPOSE 8080
CMD ["/root/cacheman"]



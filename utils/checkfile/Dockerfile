FROM golang:1.8-onbuild

MAINTAINER Anshuman Bhartiya (anshuman_bhartiya@intuit.com)

RUN go get -u github.com/daviddengcn/go-diff && go install github.com/daviddengcn/go-diff

ENTRYPOINT ["app"]




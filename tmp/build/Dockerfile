FROM alpine:3.6

ADD tmp/_output/bin/pravega-operator /usr/local/bin/pravega-operator

RUN adduser -D pravega-operator
USER pravega-operator

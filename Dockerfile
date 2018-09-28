FROM alpine:3.6

ADD /bin/pravega-operator /bin/pravega-operator

RUN adduser -D pravega-operator

USER pravega-operator

ENTRYPOINT ["/bin/pravega-operator"]

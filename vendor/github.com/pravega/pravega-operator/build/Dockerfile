#
# Copyright (c) 2017 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
FROM alpine:3.6

RUN apk add --update \
    sudo \
    libcap

ADD build/_output/bin/pravega-operator /usr/local/bin/pravega-operator
RUN sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/pravega-operator

USER nobody
FROM alpine:3.10

ENV MPY_VERSION 1.11
RUN apk add --update alpine-sdk xz python3 bsd-compat-headers libffi-dev &&\
    sed -e "s/^#warning/\/\/warning/g" -i /usr/include/sys/cdefs.h &&\
    rm /var/cache/apk/*
ADD http://micropython.org/resources/source/micropython-${MPY_VERSION}.tar.xz /src/
RUN tar -xf /src/micropython-${MPY_VERSION}.tar.xz -C /src/ &&\
    cd /src/micropython-${MPY_VERSION}/ports/unix &&\
    make coverage &&\
    mv /src/micropython-${MPY_VERSION}/ports/unix/micropython_coverage /usr/local/bin/micropython_coverage

FROM alpine:3.10
RUN apk add --update libffi && rm /var/cache/apk/*
COPY --from=0 /usr/local/bin/micropython_coverage /micropython_coverage
WORKDIR  /app

ENTRYPOINT [ "/micropython_coverage" ]
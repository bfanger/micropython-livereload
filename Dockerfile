FROM alpine:3.10

ENV MPY_VERSION 1.11
RUN apk add --update alpine-sdk xz python3 bsd-compat-headers libffi-dev &&\
    sed -e "s/^#warning/\/\/warning/g" -i /usr/include/sys/cdefs.h &&\
    rm /var/cache/apk/*
ADD http://micropython.org/resources/source/micropython-${MPY_VERSION}.tar.xz /src/
RUN tar -xf /src/micropython-${MPY_VERSION}.tar.xz -C /src/ &&\
    cd /src/micropython-${MPY_VERSION}/ports/unix &&\
    # Build micropython_coverage
    make coverage &&\
    mv /src/micropython-${MPY_VERSION}/ports/unix/micropython_coverage /bin/micropython_coverage &&\
    # Build regular micropython
    make clean && make &&\
    mv /src/micropython-${MPY_VERSION}/ports/unix/micropython /bin/micropython

FROM alpine:3.10
RUN apk add --update libffi && rm /var/cache/apk/*
COPY --from=0 /bin/micropython_coverage /bin/micropython_coverage
COPY --from=0 /bin/micropython /bin/micropython
WORKDIR  /app

ENTRYPOINT [ "/bin/micropython_coverage" ]
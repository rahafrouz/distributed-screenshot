FROM zenika/alpine-chrome


COPY worker /usr/src/app/worker
ADD https://github.com/sensepost/gowitness/releases/download/1.0.8/gowitness-linux-amd64 /usr/src/app/gowitness
USER root
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN chmod +x /usr/src/app/gowitness
RUN chown chrome:chrome /usr/src/app/*
USER root
ENTRYPOINT ["/usr/src/app/worker"]



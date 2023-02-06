FROM alpine:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
COPY caddy ./caddy
COPY helloworld ./helloworld 
COPY helloworld.json ./helloworld.json
COPY conf.sh ./conf.sh
RUN chomd +x ./conf.sh
CMD ["./conf.sh"]

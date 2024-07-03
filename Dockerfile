FROM ubuntu:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
COPY caddy ./caddy
COPY helloworld ./helloworld 
COPY helloworld.json ./helloworld.json
COPY start.sh ./start.sh
RUN chmod +x ./start.sh
CMD ["/bin/sh","./start.sh"]
EXPOSE 80

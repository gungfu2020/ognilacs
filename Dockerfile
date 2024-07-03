FROM ubuntu:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
RUN echo "nameserver 1.1.1.1" > ./etc/resolv.conf
RUN apt-get update && apt-get install -y curl
COPY caddy ./caddy
COPY helloworld ./helloworld 
COPY helloworld.json ./helloworld.json
COPY start.sh ./start.sh
RUN chmod +x ./start.sh
CMD ["/bin/sh","./start.sh"]
EXPOSE 80

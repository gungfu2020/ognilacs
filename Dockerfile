FROM ubuntu:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
RUN cat /etc/resolv.conf
RUN ls -lh /etc/resolv.conf
RUN apt-get update && apt-get install -y curl dnsutils
COPY caddy ./caddy
COPY helloworld ./helloworld 
COPY helloworld.json ./helloworld.json
COPY start.sh ./start.sh
RUN chmod +x ./start.sh
CMD ["/bin/sh","./start.sh"]
EXPOSE 80

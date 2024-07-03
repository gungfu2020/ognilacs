FROM ubuntu:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
RUN apt-get update && apt-get install -y curl
COPY caddy ./caddy
COPY helloworld ./helloworld 
COPY helloworld.json ./helloworld.json
COPY start.sh ./start.sh
RUN chmod +x ./start.sh
CMD ["/bin/sh","./start.sh"]
EXPOSE 80

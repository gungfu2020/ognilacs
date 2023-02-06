FROM alpine:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
COPY caddy ./caddy
COPY helloworld ./helloworld 
COPY helloworld.json ./helloworld.json
COPY start.sh ./start.sh
RUN chmod +x ./start.sh
CMD ["./start.sh"]
EXPOSE 8080

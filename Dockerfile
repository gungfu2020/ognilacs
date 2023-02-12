FROM alpine:latest
LABEL maintainer "niubiya <dev@niubiya.org>"
WORKDIR /
COPY pikpak-upload-server ./pikpak-upload-server
RUN chmod +x ./pikpak-upload-server
CMD ["./pikpak-upload-server"]
EXPOSE 8080

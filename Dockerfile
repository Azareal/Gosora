FROM golang:1.10.3
RUN git clone https://github.com/Azareal/Gosora
RUN mv Gosora app
ADD . /app/
WORKDIR /app
RUN ./update-deps-linux
ENTRYPOINT ["install-docker"]
CMD ["/app/run-linux"]
FROM golang:1.12.10
RUN mkdir /app
ADD . /app/
WORKDIR /app
ENV TZ=Asia/Almaty
# RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN go build
CMD ["/app/forum"]

EXPOSE 8080

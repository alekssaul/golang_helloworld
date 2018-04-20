FROM golang:latest as builder
WORKDIR /go/src/github.com/alekssaul/golang_helloworld/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hello ./server
RUN mkdir /app && \
	wget -q http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz && \
	gzip -d GeoLite2-City.mmdb.gz && \
	mv GeoLite2-City.mmdb /app && \
	mv hello /app

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app .
CMD /app/hello
EXPOSE 80

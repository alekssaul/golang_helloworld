FROM golang:latest 
RUN apt-get update
RUN rm -rf /var/lib/apt/lists/* && \
	mkdir /app && \
	go get github.com/oschwald/geoip2-golang && \
	go get github.com/kelseyhightower/envconfig
RUN wget -q http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz && \
	gzip -d GeoLite2-City.mmdb.gz && \
	mv GeoLite2-City.mmdb /app
ADD . /app/ 
WORKDIR /app 
RUN go build -o hello . 
CMD ["/app/hello"]

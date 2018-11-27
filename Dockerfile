FROM golang:alpine AS build-step
RUN apk add --no-cache git g++
ADD . /build-superman
RUN cd /build-superman && go build -o superman cmd/logins/superman.go 


FROM alpine
ENV GEOIP_BASE_URL   http://geolite.maxmind.com/download/geoip/database
ENV GEOIP_CITY_DB    GeoLite2-City
ENV GEOIP_DB_DIR     /GeoLite2
ADD ${GEOIP_BASE_URL}/${GEOIP_CITY_DB}.tar.gz /tmp/

RUN mkdir -p ${GEOIP_DB_DIR} && tar -ztf /tmp/${GEOIP_CITY_DB}.tar.gz | grep mmdb \
    | xargs -t -I {} tar -C ${GEOIP_DB_DIR} --strip-components 1 \
    -xvzf /tmp/${GEOIP_CITY_DB}.tar.gz {} \
    && rm -f /tmp/GeoLite2-*

WORKDIR /superman
COPY --from=build-step /build-superman/superman /superman/
EXPOSE 8080
ENTRYPOINT ./superman
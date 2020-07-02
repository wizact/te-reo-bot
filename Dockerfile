FROM golang:1.13-alpine AS build

RUN apk add --no-cache \
	bash \
	git \
	make

WORKDIR /build

COPY . .

RUN make build-static




FROM alpine:3.11 AS base

LABEL maintainer="amir.mohtasebi@gmail.com"

RUN set -x \
    && apk --update add ca-certificates wget gnupg && rm -rf /var/cache/apk/

EXPOSE 2020

WORKDIR /app/
COPY --from=build /build/out/ .

ENTRYPOINT ["./te-reo-bot"]
CMD ["start-server", "-address=localhost", "-port=2020", "-tls=false"]

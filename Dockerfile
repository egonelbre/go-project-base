FROM debian:wheezy

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

ENV DEVELOPMENT false

ADD . /app
WORKDIR /app/

ENV PORT 80
EXPOSE 80

RUN ["chmod", "+x", "/app/.bin/run"]
CMD ["/app/.bin/run"]
FROM golang:1.14.9-alpine as builder

RUN apk add make && \
	mkdir /project

ADD . /project/

WORKDIR /project

RUN rm -fr build \
	&& make build

FROM scratch

COPY --from=builder /project/build /

ENTRYPOINT ["/gorunner"]

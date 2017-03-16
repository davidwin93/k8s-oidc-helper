FROM alpine
RUN apk --update --no-cache add ca-certificates
COPY auth /

ENTRYPOINT ["/auth"]

FROM alpine:3.12

COPY jb /

ENTRYPOINT ["/jb"]

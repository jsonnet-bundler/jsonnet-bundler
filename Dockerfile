FROM busybox:1.35.0

COPY _output/linux/amd64/jb /

ENTRYPOINT ["/jb"]

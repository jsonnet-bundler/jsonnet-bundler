FROM busybox:1.33.0

COPY _output/linux/amd64/jb /

ENTRYPOINT ["/jb"]

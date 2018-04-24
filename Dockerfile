FROM busybox:1.28

COPY _output/linux/amd64/jb /

ENTRYPOINT ["/jb"]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.undistro.io/gcr/distroless/static:debug
WORKDIR /
COPY manager /
USER 65532:65532

ENTRYPOINT ["/manager"]

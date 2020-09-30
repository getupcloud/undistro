# Build the manager binary
FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager main.go

RUN curl -o aws-iam-authenticator https://amazon-eks.s3.us-west-2.amazonaws.com/1.17.9/2020-08-04/bin/linux/amd64/aws-iam-authenticator && chmod +x aws-iam-authenticator && mv aws-iam-authenticator /go/bin

RUN mkdir -p /app

RUN mkdir -p /home/.aws


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=builder --chown=nonroot:nonroot /app /app
COPY --from=builder --chown=nonroot:nonroot /workspace/clustertemplates /app/clustertemplates
COPY --from=builder --chown=nonroot:nonroot /home/.aws /home/.aws
COPY --from=builder /go/bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=builder /workspace/manager /app/manager
USER nonroot:nonroot
WORKDIR /app
ENTRYPOINT ["/app/manager"]

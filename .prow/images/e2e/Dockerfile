FROM golang:1.16

ENV  GO111MODULE=on

RUN apt-get update && \
    apt-get -y full-upgrade && \
    apt-get install -y curl && \
    curl -sSL https://get.docker.com/ | bash && \
    curl -LO "https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    mv kubectl /bin/kubectl && \
    go get sigs.k8s.io/kind@v0.10.0 && \
    go install github.com/goreleaser/goreleaser@v0.162.0
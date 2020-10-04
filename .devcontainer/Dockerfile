ARG VARIANT=buster
FROM mcr.microsoft.com/vscode/devcontainers/base:${VARIANT}
RUN apt update
COPY library-scripts/ /tmp/library-scripts
ENV GOROOT=/usr/local/go \
    GOPATH=/go
ENV PATH=${GOPATH}/bin:${GOROOT}/bin:${PATH}
RUN bash /tmp/library-scripts/common-debian.sh "true" "vscode" "1000" "1000" "true" \
&& bash /tmp/library-scripts/go-debian.sh "latest" "${GOROOT}" "${GOPATH}" \
&& bash /tmp/library-scripts/kubectl-helm-debian.sh \
&& bash /tmp/library-scripts/github-debian.sh \
&& bash /tmp/library-scripts/docker-debian.sh \
&& bash /tmp/library-scripts/kind-debian.sh \
&& bash /tmp/library-scripts/tilt-debian.sh \
&& apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/library-scripts
ENTRYPOINT ["/usr/local/share/docker-init.sh"]
CMD ["sleep", "infinity"]
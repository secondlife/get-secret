FROM artifactory.secondlife.io/dockerhub/golang:1.18-alpine AS build
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 go build -o get-secret

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/get-secret /usr/bin/
ENV GET_SECRETS=
ENTRYPOINT ["/usr/bin/get-secret"]
CMD ["-v", "--env-conf", "GET_SECRETS"]

ARG GOLANG_VERSION=1.15
FROM golang:${GOLANG_VERSION} as builder

WORKDIR /src

# fetch dependencies first
RUN go get -u golang.org/x/lint/golint

COPY go.mod go.sum ./
RUN go mod download -x

# copy sources
COPY . .

ARG VERSION
ARG GIT_COMMIT_SHA
RUN make VERSION=${VERSION} GIT_COMMIT_SHA=${GIT_COMMIT_SHA}

FROM golang:${GOLANG_VERSION}

COPY --from=builder /src/target/boilerplate /usr/local/bin/

ENV GIN_MODE=release PORT=80
EXPOSE ${PORT}
CMD ["boilerplate"]

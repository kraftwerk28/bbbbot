FROM golang AS build
WORKDIR /opt/build
COPY ./ ./
ARG CGO_ENABLED=0
RUN go build

FROM alpine
ARG BIN
WORKDIR /opt/app
COPY --from=build /opt/build/$BIN ./$BIN

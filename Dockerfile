FROM golang:1.21

ARG N2M_VERSION=0.0.1
RUN go install github.com/gerrowadat/nut2mqtt@${N2M_VERSION}

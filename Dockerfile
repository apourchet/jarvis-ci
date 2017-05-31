FROM golang:1.7.1 as builder
WORKDIR /go/src/github.com/apourchet/jarvis-ci 
ADD . /go/src/github.com/apourchet/jarvis-ci 
RUN CGO_ENABLED=0 go build -i -ldflags "-s" -o /jarvis-ci .

FROM jpetazzo/dind:latest as runner
RUN apt-get install -y make
COPY --from=builder /jarvis-ci /jarvis-ci
ENTRYPOINT ["wrapdocker", "/jarvis-ci"]

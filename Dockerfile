FROM golang:alpine
WORKDIR /go/src/github.com/neumayer/dbwebapp/
COPY . .
ENV CGO_ENABLED 0
RUN go build -o dbwebapp

FROM scratch
COPY --from=0 /go/src/github.com/neumayer/dbwebapp/dbwebapp /
EXPOSE 8080
ENTRYPOINT ["/dbwebapp"]

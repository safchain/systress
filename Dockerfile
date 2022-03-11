FROM golang:1.17
WORKDIR /go/src/github.com/safchain/systress
COPY . ./
RUN find 
RUN go build systress.go

FROM ubuntu  
#RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/safchain/systress/systress /
CMD ["/systress"] 

FROM  golang:1.6.2-alpine
MAINTAINER Alepht
RUN apk update && apk add git
WORKDIR /home/messenger/messenger-ssh
ENV GOPATH /home/messenger/messenger-ssh
ENV SSH_KEY $(echo $SSH_KEY)
ADD . .
RUN go get golang.org/x/crypto/ssh 
RUN go build router.go
EXPOSE 8000
ENTRYPOINT ./router
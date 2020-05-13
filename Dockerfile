FROM amazonlinux:2 as build

RUN yum update -y && yum install -y golang git && yum clean all && rm -rf /var/cache/yum
WORKDIR /root/go/src/mproxy
COPY . .
RUN go get && go install

FROM amazonlinux:2

ARG USER=mproxy

ENV HOME=/home/$USER
ENV USER_ID=1000
ENV USER_GID=100

RUN yum update -y && yum install -y shadow-utils && yum clean all && rm -rf /var/cache/yum
RUN useradd -u $USER_ID -g $USER_GID -d $HOME $USER

COPY --from=build /root/go/bin/mproxy /usr/bin/mproxy
USER $USER_ID

ENTRYPOINT ["/usr/bin/mproxy"]


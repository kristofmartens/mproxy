FROM 725028247697.dkr.ecr.eu-west-1.amazonaws.com/cicd/yad/kbc-amazon-linux as build

RUN yum update -y && yum install -y golang && yum clean all && rm -rf /var/cache/yum
WORKDIR /root/go/src/mproxy
COPY src .
RUN go install

FROM 725028247697.dkr.ecr.eu-west-1.amazonaws.com/cicd/yad/kbc-amazon-linux

ARG USER=mproxy

ENV HOME=/home/$USER
ENV USER_ID=1000
ENV USER_GID=100

COPY --from=build /root/go/bin/mproxy /usr/bin/mproxy

RUN useradd -u $USER_ID -g $USER_GID -d $HOME $USER
USER $USER_ID

ENTRYPOINT ["/usr/bin/mproxy"]


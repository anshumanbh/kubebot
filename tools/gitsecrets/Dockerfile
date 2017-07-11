FROM python:latest
MAINTAINER anshuman.bhartiya@gmail.com

RUN mkdir /data
WORKDIR /data

RUN git clone https://github.com/awslabs/git-secrets.git && cd git-secrets && make install

ADD run.sh /data
RUN chmod +x /data/run.sh
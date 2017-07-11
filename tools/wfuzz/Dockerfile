FROM ubuntu:latest
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

# Doing the usual here
RUN apt-get -y update && \
    apt-get -y dist-upgrade

RUN apt-get install -y \
	build-essential \
	git \
	libpcap-dev \
	libxml2-dev \
	libxslt1-dev \
	python-requests \
	python-dnspython \
	python-setuptools \
	python-dev \
    libcurl4-gnutls-dev \
    librtmp-dev \
	wget \
	zlib1g-dev && apt-get clean

RUN easy_install pip && pip install pycurl

RUN mkdir /data
WORKDIR /data

RUN git clone https://github.com/danielmiessler/SecLists.git
RUN git clone https://github.com/anshumanbh/wfuzz.git

WORKDIR /data/wfuzz
RUN chmod +x wfuzz.py

ENTRYPOINT ["./wfuzz.py"]
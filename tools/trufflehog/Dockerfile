FROM python:2.7
MAINTAINER Anshuman Bhartiya <anshuman.bhartiya@gmail.com>

ADD . /data
WORKDIR /data

RUN pip install -r requirements.txt
RUN chmod +x truffleHog/truffleHog.py

WORKDIR /data/truffleHog

ENTRYPOINT ["./truffleHog.py"]
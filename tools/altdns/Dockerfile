FROM python:2.7
MAINTAINER Anshuman Bhartiya <anshuman.bhartiya@gmail.com>

RUN git clone https://github.com/infosec-au/altdns.git

WORKDIR /altdns

RUN pip install -r requirements.txt

RUN wget https://raw.githubusercontent.com/anshumanbhtest/gobuster/master/gobuster_google.com

#./altdns.py -i gobuster_google.com -o data_output -w words.txt -r -e -d 8.8.8.8 -t 50 -s results.txt

ENTRYPOINT ["/bin/sh"]
FROM python:2.7
MAINTAINER Anshuman Bhartiya <anshuman.bhartiya@gmail.com>

RUN git clone https://github.com/TheRook/subbrute.git

WORKDIR /subbrute

RUN mkdir subfiles && cd subfiles && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/bitquark_20160227_subdomains_popular_1000000.txt && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/deepmagic.com_top500prefixes.txt && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/fierce_hostlist.txt && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/namelist.txt && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/names.txt && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/sorted_knock_dnsrecon_fierce_recon-ng.txt && \
    wget https://raw.githubusercontent.com/anshumanbh/brutesubs/master/wordlists/subdomains-top1mil-110000.txt

ENTRYPOINT ["./subbrute.py"]
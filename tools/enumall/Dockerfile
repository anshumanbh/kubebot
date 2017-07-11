FROM python:2.7
MAINTAINER Anshuman Bhartiya <anshuman.bhartiya@gmail.com>

RUN git clone https://LaNMaSteR53@bitbucket.org/LaNMaSteR53/recon-ng.git

WORKDIR /recon-ng
COPY enumall-ab.py .
RUN chmod +x enumall-ab.py

RUN pip install -r REQUIREMENTS && ln -s /recon-ng /usr/share/recon-ng

ENTRYPOINT ["./enumall-ab.py"]
FROM kalilinux/kali-linux-docker
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

RUN apt-get -y update && apt-get -y dist-upgrade && apt-get clean
RUN apt-get install -y metasploit-framework

CMD ["/bin/bash"]
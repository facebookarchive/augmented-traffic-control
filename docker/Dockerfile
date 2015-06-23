FROM python:2.7-onbuild
MAINTAINER Emmanuel Bretelle <chantra@fb.com>

# Assuming the default WAN and LAN interface are eth0 and eth1
ENV ATCD_WAN=eth0 ATCD_LAN=eth1 ATCD_BURST_SIZE=12000 ATCD_MODE=secure
RUN apt-get update && apt-get install -y iptables rsyslog
RUN django-admin startproject atcui . || true
RUN cp -f atcui_tpl/* atcui/
RUN python manage.py migrate


# Supervisor setup
ADD ./start.sh /start.sh
RUN chmod 755 /start.sh

EXPOSE 8000

CMD ["/bin/bash", "/start.sh"]

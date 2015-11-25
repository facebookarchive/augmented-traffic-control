FROM golang:1.5.1
RUN apt-get -y update; apt-get -y install iptables rsyslog; systemctl enable rsyslog; apt-get -y clean

COPY ./utils/docker_entrypoint.sh /
ENTRYPOINT ["/docker_entrypoint.sh"]
CMD ["bash"]

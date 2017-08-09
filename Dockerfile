FROM golang:latest
RUN apt-get -y update && apt-get -y install iptables rsyslog iperf3 && systemctl enable rsyslog && apt-get -y clean

COPY ./utils/docker_entrypoint.sh /
ENTRYPOINT ["/docker_entrypoint.sh"]
CMD ["bash"]

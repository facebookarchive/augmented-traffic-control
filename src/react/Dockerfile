FROM node:latest
WORKDIR /usr/src/myapp/src/react
RUN echo 'export PATH=/usr/src/myapp/src/react/node_modules/.bin:$PATH' > /etc/profile.d/node.sh
COPY ./docker_entrypoint.sh /
ENTRYPOINT ["/docker_entrypoint.sh"]
CMD ["bash"]

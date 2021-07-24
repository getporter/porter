FROM gitpod/workspace-full

ENV TEMP_DEB="$(mktemp)" 
RUN wget -O "$TEMP_DEB" 'https://github.com/gohugoio/hugo/releases/download/v0.86.0/hugo_0.86.0_Linux-64bit.deb'
RUN sudo dpkg -i "$TEMP_DEB"
RUN rm -f "$TEMP_DEB"
ENV TEMP_DEB=
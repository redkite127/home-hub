FROM alpine:3.6

ENV bin_dir /opt/home-hub/bin
ENV etc_dir /opt/home-hub/etc
ENV var_dir /opt/home-hub/var

RUN mkdir -p ${bin_dir} && mkdir -p ${etc_dir} && mkdir -p ${var_dir}

COPY home-hub ${bin_dir}/home-hub

RUN chmod +x ${bin_dir}/home-hub

WORKDIR ${bin_dir}

# it does accept the variable ${etc_dir} in the parameters
#CMD ["./home-hub", "-config-dir", "/opt/tadaweb/etc"]
CMD ["./home-hub"]

FROM flant/shell-operator:latest

ADD https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz /opt/oc/release.tar.gz
RUN apk add --no-cache ca-certificates && tar --strip-components=1 -xzvf  /opt/oc/release.tar.gz -C /opt/oc/ && \
    mv /opt/oc/oc /usr/local/bin/ && \
    rm -rf /opt/oc
ADD ./hooks/* /hooks/
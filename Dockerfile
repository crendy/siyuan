#FROM node:16 as NODE_BUILD
FROM registry.cn-hangzhou.aliyuncs.com/crendy/node:21 as NODE_BUILD
WORKDIR /go/src/github.com/siyuan-note/siyuan/
ADD . /go/src/github.com/siyuan-note/siyuan/
RUN apt-get update && \
    apt-get install -y jq
RUN cd app && \
packageManager=$(jq -r '.packageManager' package.json) && \
if [ -n "$packageManager" ]; then \
    npm install -g $packageManager; \
else \
    echo "No packageManager field found in package.json"; \
    npm install -g pnpm; \
fi && \
pnpm install --registry=http://registry.npmjs.org/ --silent && \
pnpm run build
RUN apt-get purge -y jq
RUN apt-get autoremove -y
RUN rm -rf /var/lib/apt/lists/*

#FROM golang:alpine as GO_BUILD
FROM registry.cn-hangzhou.aliyuncs.com/crendy/golang:1.22.0-alpine as GO_BUILD
WORKDIR /go/src/github.com/siyuan-note/siyuan/
COPY --from=NODE_BUILD /go/src/github.com/siyuan-note/siyuan/ /go/src/github.com/siyuan-note/siyuan/
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev && \
    cd kernel && go build --tags fts5 -v -ldflags "-s -w" && \
    mkdir /opt/siyuan/ && \
    mv /go/src/github.com/siyuan-note/siyuan/app/appearance/ /opt/siyuan/ && \
    mv /go/src/github.com/siyuan-note/siyuan/app/stage/ /opt/siyuan/ && \
    mv /go/src/github.com/siyuan-note/siyuan/app/guide/ /opt/siyuan/ && \
    mv /go/src/github.com/siyuan-note/siyuan/app/changelogs/ /opt/siyuan/ && \
    mv /go/src/github.com/siyuan-note/siyuan/kernel/kernel /opt/siyuan/ && \
    mv /go/src/github.com/siyuan-note/siyuan/kernel/entrypoint.sh /opt/siyuan/entrypoint.sh && \
    find /opt/siyuan/ -name .git | xargs rm -rf

FROM registry.cn-hangzhou.aliyuncs.com/crendy/alpine:latest
LABEL maintainer="crendy<crendy@126.com>"

WORKDIR /opt/siyuan/
COPY --from=GO_BUILD /opt/siyuan/ /opt/siyuan/
#RUN addgroup --gid 1000 siyuan && adduser --uid 1000 --ingroup siyuan --disabled-password siyuan && apk add --no-cache ca-certificates tzdata && chown -R siyuan:siyuan /opt/siyuan/

#ENV TZ=Asia/Shanghai
#ENV RUN_IN_CONTAINER=true
#EXPOSE 6806

#USER siyuan
#ENTRYPOINT ["/opt/siyuan/kernel"]
RUN \
  apk add --no-cache \
    ca-certificates \
    tzdata && \
  echo "**** cleanup ****" && \
  rm -rf \
      /root/.cache \
      /tmp/*

ENV TZ=Asia/Shanghai
ENV WORKSPACE=/root/SiYuan
ENV ACCESSAUTHCODE=
ENV PUID=0
ENV PGID=0
ENV RUN_IN_CONTAINER=true
VOLUME /root/SiYuan
EXPOSE 6806

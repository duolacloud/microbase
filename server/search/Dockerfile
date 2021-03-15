FROM alpine:3.10
ADD build/search-srv-linux /search-srv
## 解决alipay获取时区错误
ENV ZONEINFO /opt/zoneinfo.zip

ENTRYPOINT ["/search-srv"]
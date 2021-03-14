### elastic 分页问题

#### elastic order

排序的字段需要 mapping 设置 fielddata: true, 否则会报错

#### connection 分页实现

#### 中文搜索问题

#### 全词匹配问题

#### index struct tag

搜索下 golang elasticsearch orm 相关

### 云原生 搜索服务实现

云原生策略的话，应该是
frontend --- https/wss/json ---> api --- grpc ---> biz.service --- grpc ---> search.service --- http ---> elasticsearch

参考:
product graphql
https://github.com/jacob-ebey/golang-ecomm/blob/9e88e660c2d435c248c073a9892c5ad37db9605d/dataloaders/product.go

gqlgen-authentication
https://github.com/AneriShah2610/gqlgen-authentication

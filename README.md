# Tracer-demo 链路测试
基于`alibaba`开源的 [opentelemetry-go-auto-instrumentation](github.com/alibaba/opentelemetry-go-auto-instrumentation) 实现`Golang`无侵入式注入链路追踪能力

可以运行多个`Pod`来实现简单全链路追踪

## 核心环境变量配置
``` 
            - name: OTEL_EXPORTER_OTLP_ENDPOINT
              value: http://otel-collector.observability:4318
            - name: SERVICE_NAME
              value: "tracer-a"
            - name: DOWNSTREAM_SERVICE_URL  // 下游
              value: "http://tracer-b:8080"
```
**解释**
* `SERVICE_NAME`：服务名称，用于体现在链路当中，能够明确定位该服务
* `DOWNSTREAM_SERVICE_URL`：下游服务地址，可以运行多个`Pod`分别命名为不同的名称，如果没有配置或者为空，说明该服务是没有下游的最终节点

## 部署应用

[部署文件](./deploy.yaml)

```
# ls
a.yaml  b.yaml

# cat a.yaml
...
            - name: SERVICE_NAME
              value: "tracer-a"
            - name: DOWNSTREAM_SERVICE_URL
              value: "http://tracer-b:8080"
...

# cat b.yaml
...
            - name: SERVICE_NAME
              value: "tracer-b"
            - name: DOWNSTREAM_SERVICE_URL
              value: ""
...

# kubectl get po -o wide
NAME                                     READY   STATUS    RESTARTS       AGE    IP            NODE      NOMINATED NODE   READINESS GATES
tracer-b-5bc4bbb6cc-wnthn                1/1     Running   0              35m    10.42.0.195   master1   <none>           <none>
tracer-a-77b9556f74-lvkjb                1/1     Running   0              35m    10.42.0.194   master1   <none>           <none>
```

## 验证链路
``` 
# curl -H 'baggage: name=cairry' 10.42.0.194:8080/call
{"service":"tracer-b","message":"no downstream service","at":"2025-09-12T08:42:23.281264663Z","traceparent":"00-0ca05b8b3a7e55069793b609fcb4c939-00f0280f8a74926f-01"}
```
> 从`tracer-a`服务发起，最终节点是`tracer-b`

**Tracer-a 日志**
``` 
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=fe1e5582bf1749e3[tracer-a] incoming request GET /call traceparent=
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=fe1e5582bf1749e3[tracer-a] request headers: map[Accept:[*/*] Baggage:[name=cairry] User-Agent:[curl/7.29.0]]
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=fe1e5582bf1749e3[tracer-a] completed http://tracer-b:8080 in 2.97442ms traceparent=
```
**Tracer-b 日志**
``` 
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=2ea2520142d12f05[tracer-b] incoming request GET /call traceparent=00-0ca05b8b3a7e55069793b609fcb4c939-00f0280f8a74926f-01
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=2ea2520142d12f05[tracer-b] request headers: map[Accept-Encoding:[gzip] Baggage:[name=cairry] Traceparent:[00-0ca05b8b3a7e55069793b609fcb4c939-00f0280f8a74926f-01] User-Agent:[Go-http-client/1.1]]
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=2ea2520142d12f05[tracer-b] no downstream service configured, skipping call
2025/09/12 08:42:23  trace_id=0ca05b8b3a7e55069793b609fcb4c939 span_id=2ea2520142d12f05[tracer-b] completed  in 56.76µs traceparent=00-0ca05b8b3a7e55069793b609fcb4c939-00f0280f8a74926f-01
```
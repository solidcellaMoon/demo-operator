```bash
❯ go run main.go
2022-04-23T17:07:57.893+0900    INFO    controller-runtime.metrics      metrics server is starting to listen    {"addr": ":8080"}
2022-04-23T17:07:57.893+0900    INFO    setup   starting manager
2022-04-23T17:07:57.893+0900    INFO    starting metrics server {"path": "/metrics"}
2022-04-23T17:07:57.893+0900    INFO    controller.demo Starting EventSource    {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo", "source": "kind source: /, Kind="}
2022-04-23T17:07:57.894+0900    INFO    controller.demo Starting EventSource    {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo", "source": "kind source: /, Kind="}
2022-04-23T17:07:57.894+0900    INFO    controller.demo Starting EventSource    {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo", "source": "kind source: /, Kind="}
2022-04-23T17:07:57.894+0900    INFO    controller.demo Starting Controller     {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo"}
2022-04-23T17:07:57.994+0900    INFO    controller.demo Starting workers        {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo", "worker count": 1}
# demo CR 적용 이후 로그
# .
# .
# .
2022-04-23T17:08:35.101+0900    INFO    Resource Changed
2022-04-23T17:08:35.112+0900    INFO    Service Created {"svc.namespace": "default", "svc.name": "demo-sample"}
2022-04-23T17:08:35.112+0900    INFO    Resource Changed
2022-04-23T17:08:35.119+0900    INFO    Service Created {"deploy.namespace": "default", "deploy.name": "demo-sample"}
2022-04-23T17:08:35.119+0900    INFO    Resource Changed
2022-04-23T17:08:35.131+0900    INFO    Resource Changed
2022-04-23T17:08:35.149+0900    INFO    Resource Changed
2022-04-23T17:08:35.161+0900    INFO    Resource Changed
2022-04-23T17:08:37.113+0900    INFO    Resource Changed
2022-04-23T17:08:46.954+0900    INFO    Resource Changed
# size 1 -> 2
# .
# .
# .
2022-04-23T17:18:52.361+0900    INFO    Resource Changed
2022-04-23T17:18:52.361+0900    INFO    changed replicas Size   {"ReplicaSet.namespace": "default", "Size": 2}
2022-04-23T17:18:52.367+0900    INFO    Resource Changed
2022-04-23T17:18:52.382+0900    INFO    Resource Changed
2022-04-23T17:18:52.392+0900    INFO    Resource Changed
2022-04-23T17:18:52.407+0900    INFO    Resource Changed
2022-04-23T17:18:54.367+0900    INFO    Resource Changed
2022-04-23T17:18:56.797+0900    INFO    Resource Changed
# deploy demo-sample 삭제
# .
# .
# .
2022-04-23T17:22:13.106+0900    INFO    Resource Changed
2022-04-23T17:22:13.111+0900    INFO    Service Created {"deploy.namespace": "default", "deploy.name": "demo-sample"}
2022-04-23T17:22:13.111+0900    INFO    Resource Changed
2022-04-23T17:22:13.127+0900    INFO    Resource Changed
2022-04-23T17:22:13.140+0900    INFO    Resource Changed
2022-04-23T17:22:13.164+0900    INFO    Resource Changed
2022-04-23T17:22:13.189+0900    INFO    Resource Changed
2022-04-23T17:22:15.112+0900    INFO    Resource Changed
2022-04-23T17:22:18.820+0900    INFO    Resource Changed
2022-04-23T17:22:20.841+0900    INFO    Resource Changed
2022-04-23T17:23:10.373+0900    INFO    Resource Changed
# service demo-sample 삭제
# .
# .
# .
2022-04-23T17:24:02.208+0900    INFO    Resource Changed
2022-04-23T17:24:02.222+0900    INFO    Service Created {"svc.namespace": "default", "svc.name": "demo-sample"}
2022-04-23T17:24:02.222+0900    INFO    Resource Changed
2022-04-23T17:24:04.223+0900    INFO    Resource Changed
# demo CR 삭제 테스트
# .
# .
# .
2022-04-23T17:26:53.112+0900    INFO    Resource Changed
2022-04-23T17:26:53.112+0900    INFO    Deleted - Demo CR
2022-04-23T17:26:53.145+0900    INFO    Resource Changed
2022-04-23T17:26:53.145+0900    INFO    Deleted - Demo CR
2022-04-23T17:26:53.148+0900    INFO    Resource Changed
2022-04-23T17:26:53.148+0900    INFO    Deleted - Demo CR
# 재생성
# .
# .
# .
2022-04-23T17:29:25.186+0900    INFO    Resource Changed
2022-04-23T17:29:25.196+0900    INFO    Service Created {"svc.namespace": "default", "svc.name": "demo-sample"}
2022-04-23T17:29:25.196+0900    INFO    Resource Changed
2022-04-23T17:29:25.200+0900    INFO    Service Created {"deploy.namespace": "default", "deploy.name": "demo-sample"}
2022-04-23T17:29:25.200+0900    INFO    Resource Changed
2022-04-23T17:29:25.221+0900    INFO    Resource Changed
2022-04-23T17:29:25.233+0900    INFO    Resource Changed
2022-04-23T17:29:25.257+0900    INFO    Resource Changed
2022-04-23T17:29:27.196+0900    INFO    Resource Changed
2022-04-23T17:29:30.235+0900    INFO    Resource Changed
2022-04-23T17:29:32.255+0900    INFO    Resource Changed
# cr 생성 확인 이후 operator 빌드를 꺼봅니다.
# .
# .
# .
^C2022-04-23T17:30:43.818+0900  INFO    controller.demo Shutdown signal received, waiting for all workers to finish    {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo"}
2022-04-23T17:30:43.818+0900    INFO    controller.demo All workers finished    {"reconciler group": "demoapp.my.domain", "reconciler kind": "Demo"}
```
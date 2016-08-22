# capman
A manager for CAP systems to track peers and event on leave/join

## Usage
Heartbeat based on health check running scripts for various events as other peers report in
```
capman --host=$(hostname) \
  --ip=${LOCAL_IPV4} \
  --prefix=k8s/master/runtime/${CLOUD_DETAIL}/etcd \
  heartbeat \
  --liveliness-check-url=http://localhost:4001 \
  --liveliness-check-timeout=10s \
  --bootstrap="/opt/etcd/bootstrap.sh" \
  --quorum-gained="/opt/etcd/quorum-gained.sh" \
  --quorum-lost="/opt/etcd/quorum-lost.sh" \
  --peer-join="/opt/etcd/peer-join.sh" \
  --peer-leave="/opt/etcd/peer-leave.sh"
```

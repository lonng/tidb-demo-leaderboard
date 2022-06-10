# tidb-demo-trending
Demonstrate how to use TiCDC continuously synchronize TiDB change to Redis

## Get Start

1. Install TiUP `curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh`
2. Run TiDB cluster `tiup -T tidb-demo-trending playground --tiflash 0 --ticdc 1`
3. Install Redis reference [Redis Get Start](https://redis.io/docs/getting-started/) and run redis `redis-server`
4. Install Kafka reference [Kafka Quickstart](https://kafka.apache.org/quickstart#quickstart_startserver)
5. Create changefeed `tiup cdc cli changefeed create --sink-uri="kafka://127.0.0.1:9092/trending?kafka-version=2.13.0&partition-num=1&max-message-bytes=67108864&replication-factor=1&protocol=canal-json"`
CREATE DATABASE IF NOT EXISTS clash;

CREATE TABLE IF NOT EXISTS clash.traffic_total
(
    connectionSize Int32,
    downloadTotal  Int64,
    uploadTotal    Int64,
    addTime        DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY addTime
      TTL addTime + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS clash.traffic
(
    down    Int64,
    up      Int64,
    addTime DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY addTime
      TTL addTime + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS clash.proxy_dial
(
    type     String,
    address  String,
    chain    Array(String),
    duration Int64,
    host     String,
    id       UUID,
    proxy    String,
    addTime  DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY (address, host, proxy)
      TTL addTime + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS clash.rule_match
(
    type     String,
    duration Int64,
    id       UUID,
    metadata Map(String, String),
    payload  String,
    proxy    String,
    rule     String,
    addTime  DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY (proxy, rule)
      TTL addTime + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS clash.dns_request
(
    answer   Array(String),
    dnsType  String,
    duration Int64,
    id       UUID,
    name     String,
    qType    String,
    source   String,
    type     String,
    addTime  DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY (name, qType, source, answer);

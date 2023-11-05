create database if not exists clash;

CREATE TABLE IF NOT EXISTS traffic_total
(
    connectionSize Int32,
    downloadTotal  Int64,
    uploadTotal    Int64,
    addTime        DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY addTime;

CREATE TABLE IF NOT EXISTS traffic
(
    down    Int64,
    up      Int64,
    addTime DateTime DEFAULT now()
) ENGINE = MergeTree()
      ORDER BY addTime;

CREATE TABLE IF NOT EXISTS proxy_dial
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
      ORDER BY (address, host, proxy);

CREATE TABLE IF NOT EXISTS rule_match
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
      ORDER BY (proxy, rule);

CREATE TABLE IF NOT EXISTS dns_request
(
    answer   Array(String),
    dnsType  String,
    duration Int64,
    id       UUID,
    name     String,
    qType    String,
    source   String,
    type     String
) ENGINE = MergeTree()
      ORDER BY (name, qType, source, answer);

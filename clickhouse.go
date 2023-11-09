package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ClickHouse/clickhouse-go/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var client *sql.DB

func init() {
	clickhouseAddr := os.Getenv("CLICKHOUSE_ADDR")
	database := os.Getenv("CLICKHOUSE_DATABASE")
	username := os.Getenv("CLICKHOUSE_USERNAME")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	clickhouseClient, err := InitClickhouse(clickhouseAddr, database, username, password)
	if err != nil {
		log.Error("初始化 Clickhouse 失败: ", err)
	}
	client = clickhouseClient
}

func InitClickhouse(
	clickhouseAddr string,
	database string,
	username string,
	password string,
) (*sql.DB, error) {

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{clickhouseAddr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
		Protocol:    clickhouse.Native,
	})

	if err := conn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Fatalf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}

func HandleMetricsDataByClickhouse(typeName string, data []byte) error {
	var tempObj = map[string]interface{}{}
	err := json.Unmarshal(data, &tempObj)
	if err != nil {
		log.Error("反序列化日志错误：", err)
		return err
	}

	switch typeName {
	case "Traffic":
		saveTraffic(tempObj)
		break
	case "TrafficTotal":
		saveTrafficTotal(tempObj)
		break
	case "RuleMatch":
		saveRuleMatch(tempObj)
		break
	case "ProxyDial":
		saveProxyDial(tempObj)
		break
	case "DNSRequest":
		saveDnsRequest(tempObj)
		break
	default:
		log.Error("未知的数据类型: ", typeName)
		return errors.New("未知的数据类型")
	}
	return nil
}

func saveDnsRequest(data map[string]interface{}) {
	ctx := clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))
	_, err := client.ExecContext(ctx,
		`INSERT INTO dns_request (answer, dnsType, duration, id, name, qType, source, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		data["answer"], data["dnsType"], data["duration"], data["id"], data["name"], data["qType"], data["source"], data["type"])

	if err != nil {
		log.Error("保存 dns_request 数据失败: ", err)
	}
}

func saveProxyDial(data map[string]interface{}) {
	ctx := clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))

	_, err := client.ExecContext(ctx, `
    INSERT INTO clash.proxy_dial (type, address, chain, duration, host, id, proxy)
    VALUES (?, ?, ?, ?, ?, ?, ?)
`, data["type"], data["address"], data["chain"], data["duration"], data["host"], data["id"], data["proxy"])

	if err != nil {
		log.Error("保存 proxy_dial 数据失败: ", err)
	}
}

func saveRuleMatch(data map[string]interface{}) {
	ctx := clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))

	_, err := client.ExecContext(ctx, `INSERT INTO rule_match (type, duration, id, metadata, payload, proxy, rule)
    VALUES (?, ?, ?, ?, ?, ?, ?)
`, data["type"], data["duration"], data["id"], data["metadata"], data["payload"], data["proxy"], data["rule"])
	if err != nil {
		log.Error("保存 rule_match 数据失败: ", err)
	}

}

func saveTrafficTotal(data map[string]interface{}) {
	ctx := clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))

	_, err := client.ExecContext(ctx, `INSERT INTO traffic_total (connectionSize, downloadTotal, uploadTotal) VALUES (?, ?, ?)`,
		data["connectionSize"], data["downloadTotal"], data["uploadTotal"])

	if err != nil {
		log.Error("保存 traffic_total 数据失败: ", err)
	}
}

func saveTraffic(data map[string]interface{}) {
	ctx := clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))
	_, err := client.ExecContext(ctx, `INSERT INTO clash.traffic (up, down)VALUES (
				?, ?
			)`, data["up"], data["down"])

	if err != nil {

	}
}

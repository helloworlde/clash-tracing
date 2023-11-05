package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/avast/retry-go/v4"
	"github.com/grafana/loki-client-go/loki"
)

func handleReport(lokiAddr, clashHost, clashToken string) {
	var client *loki.Client
	for {
		retry.Do(
			func() (err error) {
				client, err = InitClient(lokiAddr)
				return
			},
			retry.Attempts(0),
			retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
				max := time.Duration(n)
				if max > 8 {
					max = 8
				}
				duration := time.Second * max * max
				log.Errorf("第 %d 次连接 '%s' 失败，错误信息: '%v'，请检查地址或密码是否正确或者 Clash 是否开启 tracing", n, lokiAddr, err.Error())
				return duration
			}),
			retry.MaxDelay(time.Second*64),
		)

		log.Infof("连接地址 %s 成功, 开始推送数据到 Loki\n", lokiAddr)
		handleTCPConn(client, clashHost, clashToken)
	}
}

func handleTCPConn(client *loki.Client, clashHost string, clashToken string) {
	trafficCh := make(chan []byte)
	tracingCh := make(chan []byte)
	connectionCh := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	trafficDone := dialTrafficChan(ctx, clashHost, clashToken, trafficCh)
	tracingDone := dialTracingChan(ctx, clashHost, clashToken, tracingCh)
	connectionDone := dialConnectionChan(ctx, clashHost, clashToken, connectionCh)
Out:
	for {
		var buf []byte
		select {
		case buf = <-trafficCh:
		case buf = <-tracingCh:
		case buf = <-connectionCh:
		}
		typeName, content, err := HandleMetricsData(buf)

		if err != nil {
			log.Error("解析数据错误: ", err)
			break Out
		}
		//err = PushToLoki(client, typeName, content)
		//if err != nil {
		//	log.Error("推送日志到 Loki 错误: ", err)
		//	break Out
		//}

		err = HandleMetricsDataByClickhouse(typeName, content)
		if err != nil {
			log.Error("推送日志到 Clickhouse 错误: ", err)
		}
	}

	cancel()

	go func() {
		for range trafficCh {
		}
		for range tracingCh {
		}
		for range connectionCh {
		}
	}()

	<-trafficDone
	<-tracingDone
	<-connectionDone
}

func dialTrafficChan(ctx context.Context, clashHost string, clashToken string, ch chan []byte) chan struct{} {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/traffic", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/traffic?token=%s", clashHost, clashToken)
	}

	return dialWebsocketToChan(context.Background(), clashUrl, ch)
}

func dialTracingChan(ctx context.Context, clashHost string, clashToken string, ch chan []byte) chan struct{} {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/profile/tracing", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/profile/tracing?token=%s", clashHost, clashToken)
	}

	return dialWebsocketToChan(context.Background(), clashUrl, ch)
}

func dialConnectionChan(ctx context.Context, clashHost string, clashToken string, ch chan []byte) chan struct{} {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/connections", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/connections?token=%s", clashHost, clashToken)
	}

	return dialWebsocketToChan(context.Background(), clashUrl, ch)
}

func HandleMetricsData(data []byte) (string, []byte, error) {
	var tempObj = map[string]interface{}{}
	err := json.Unmarshal(data, &tempObj)
	if err != nil {
		log.Error("反序列化日志错误：", err)
		return "", nil, err
	}

	var typeName string
	if tempObj["up"] != nil {
		typeName = "Traffic"
	} else if tempObj["connections"] != nil {
		// 将 connections 替换为 traffic_total
		typeName = "TrafficTotal"
		// connection 信息在 tracing 中已经有了，所以直接删掉，只保留连接数量信息
		connections := tempObj["connections"]
		connectionsSlices := connections.([]interface{})
		tempObj["connectionSize"] = len(connectionsSlices)
		delete(tempObj, "connections")
	} else {
		typeName = fmt.Sprintf("%s", tempObj["type"])
	}
	contentBytes, err := json.Marshal(tempObj)
	if err != nil {
		log.Error("序列化处理后的数据失败: ", err)
		return "", nil, err
	}
	return typeName, contentBytes, nil
}

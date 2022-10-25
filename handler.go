package main

import (
	"context"
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
		if err := HandleMetricsData(client, buf); err != nil {
			log.Error("推送日志到 Loki 错误: ", err)
			break Out
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

package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/grafana/loki-client-go/loki"
)

func handleReport(lokiAddr, clashHost, clashToken string) {
	var client *loki.Client
	for {
		var conn net.Conn
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
				fmt.Printf("dial %s failed %d times: %v, wait %s\n", lokiAddr, n, err, duration.String())
				return duration
			}),
			retry.MaxDelay(time.Second*64),
		)

		fmt.Printf("Dial %s success, send data to LOKI\n", lokiAddr)
		handleTCPConn(client, clashHost, clashToken)

		conn.Close()
	}
}

func handleTCPConn(client *loki.Client, clashHost string, clashToken string) {
	trafficCh := make(chan []byte)
	tracingCh := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	trafficDone := dialTrafficChan(ctx, clashHost, clashToken, trafficCh)
	tracingDone := dialTracingChan(ctx, clashHost, clashToken, tracingCh)

Out:
	for {
		var buf []byte
		select {
		case buf = <-trafficCh:
		case buf = <-tracingCh:
		}
		if err := WriteToLoki(client, buf); err != nil {
			break Out
		}
	}

	cancel()

	go func() {
		for range trafficCh {
		}
		for range tracingCh {
		}
	}()

	<-trafficDone
	<-tracingDone
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

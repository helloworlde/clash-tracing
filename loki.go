package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/grafana/loki-client-go/loki"
	"github.com/grafana/loki-client-go/pkg/urlutil"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
)

func InitClient(lokiAddr string) (*loki.Client, error) {
	netUrl, err := url.Parse(lokiAddr)
	if err != nil {
		return nil, errors.New("解析 URL " + lokiAddr + "失败")
	}

	config := loki.Config{
		URL: urlutil.URLValue{
			URL: netUrl,
		},
		Timeout: 5 * time.Second,
	}

	client, err := loki.New(config)
	if err != nil {
		return nil, errors.New("初始化 Loki Client 失败")
	}
	return client, nil
}

func WriteToLoki(client *loki.Client, data []byte) error {
	content := string(data)
	var tempObj = map[string]interface{}{}
	err := json.Unmarshal(data, &tempObj)
	if err != nil {
		log.Error("反序列化日志错误：", err)
		return err
	}

	var typeName string
	if tempObj["up"] != nil {
		typeName = "traffic"
	} else if tempObj["connections"] != nil {
		typeName = "connection"
	} else {
		typeName = strings.ToLower(fmt.Sprintf("%s", tempObj["type"]))
	}

	labels := model.LabelSet{
		"job":  model.LabelValue("clash"),
		"type": model.LabelValue(typeName),
	}
	log.Debugf("类型: %s, 内容: %s", typeName, content)
	err = client.Handle(labels, time.Now(), content)
	if err != nil {
		log.Error("发送日志失败：", err)
		return err
	}
	return nil
}

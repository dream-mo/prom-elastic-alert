package alertmanager

import (
	"fmt"
	"github.com/dream-mo/prom-elastic-alert/utils/logger"
	"net/http"
	"strings"
	"time"
)

const (
	HttpAlertmanagerTimeout = 10
)

// HttpSendAlert is responsible for sending alarm content to the alertmanager via http post
func HttpSendAlert(url string, username string, password string, payload string) (bool, int) {
	requests := http.DefaultClient
	requests.Timeout = time.Second * HttpAlertmanagerTimeout
	contentType := "application/json"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		t := fmt.Sprintf("http.NewRequest error: %s", err.Error())
		logger.Logger.Errorln(t)
	}
	req.Header.Set("Content-Type", contentType)
	resp, e := requests.Do(req)
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	if e != nil {
		t := fmt.Sprintf("send alert to alertmanager error: %s", e.Error())
		logger.Logger.Errorln(t)
		if resp != nil {
			return false, resp.StatusCode
		} else {
			return false, 499
		}
	} else {
		t := fmt.Sprintf("status:%d payload:%s", resp.StatusCode, payload)
		logger.Logger.Debugln(t)
		return true, resp.StatusCode
	}
}

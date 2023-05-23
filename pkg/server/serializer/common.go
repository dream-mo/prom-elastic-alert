package serializer

import (
	"github.com/gin-gonic/gin"
	logger "github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	"net/http"
)

const (
	EnvironmentCodeNull     = ""
	EndpointFoundRecordNull = "Error 1146: Table 'gateway_admin.rows' doesn't exist"
)

type SuccessResModel struct {
	Data interface{} `json:"data,omitempty"`
}

type ResPageDataModel struct {
	Data  interface{} `json:"data,omitempty"`
	Page  int         `json:"page"`
	Total int64       `json:"total"`
}

type ErrorResModel struct {
	Status           ErrorStr          `json:"status"`
	Message          string            `json:"message"`
	LocalizedMessage *LocalizedMessage `json:"localized_message,omitempty"`
}

type LocalizedMessage struct {
	Locale  string `json:"locale"`
	Message string `json:"message"`
}

func SuccessRes(c *gin.Context, msg string) {
	logger.Logger.Infof(msg)
	c.JSON(http.StatusOK, nil)
	c.Abort()
}

func SuccessDataRes(c *gin.Context, data interface{}, msg string) {
	logger.Logger.Infof(msg)
	c.JSON(http.StatusOK, SuccessResModel{Data: data})
	c.Abort()
}

func SuccessPageDataRes(c *gin.Context, data interface{}, page int, count int64, msg string) {
	logger.Logger.Infof(msg)
	c.JSON(http.StatusOK, ResPageDataModel{Data: data, Page: page, Total: count})
	c.Abort()
}

func ErrorRes(c *gin.Context, status ErrorStr, msg string) {
	logger.Logger.Errorf(msg)
	c.JSON(ErrCodeMap[status], ErrorResModel{Status: status, Message: msg})
	c.Abort()
}

func ErrorLocalRes(c *gin.Context, status ErrorStr, msg string, locale string, localMessage string) {
	logger.Logger.Errorf(msg)
	data := ErrorResModel{
		Status:  status,
		Message: msg,
		LocalizedMessage: &LocalizedMessage{
			Locale:  locale,
			Message: localMessage,
		},
	}
	c.JSON(ErrCodeMap[status], data)
	c.Abort()
}

type PageParam struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

type EndpointParam struct {
	RouteId         int
	PolicyId        int
	KeyId           int
	CorsId          int
	Method          string
	Keyword         string
	EnvironmentCode string
	ApiType         string
}

type KeyParam struct {
	EnvironmentCode string
	Name            string
	KeyType         string
}

type PolicyParam struct {
	EnvironmentCode string
	Name            string
	PolicyType      string
}

type RouteParam struct {
	EnvironmentCode string
	Keyword         string
	Strategy        string
}

type AlarmParam struct {
	SmsType string
}

type NotifyChannelQuery struct {
	NotifyName         string
	NotifyType         string
	NotifyConfigTypeId string
}

type AlarmRuleQuery struct {
	RuleAlias string
	NotifyId  string
}

type AlarmRecordQuery struct {
	Begin      string `json:"begin"`
	End        string `json:"end"`
	RecordName string `json:"record_name"`
}

type AllWhiteIps struct {
	EnvironmentCode string `json:"environment_code"`
	IpAddress       string `json:"ip_address"`
}

func (AllWhiteIps) TableName() string {
	return "t_white_ip"
}

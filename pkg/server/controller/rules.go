package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/openinsight-proj/elastic-alert/pkg/server/serializer"
)

func FindAllRules(c *gin.Context) {

	serializer.SuccessDataRes(c, "hello", fmt.Sprintf("get all rules success"))
}

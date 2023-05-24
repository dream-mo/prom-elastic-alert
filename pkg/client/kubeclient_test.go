package client

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListConfigmaps(t *testing.T) {
	c, err := GetClientSet()
	if err != nil {
		panic(err)
	}

	cms, err := c.CoreV1().ConfigMaps("insight-system").Get(context.Background(), "elastic-alert-rules", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	for key, val := range cms.Data {
		fmt.Printf("cm: %s, v: %s", key, val)
	}

}

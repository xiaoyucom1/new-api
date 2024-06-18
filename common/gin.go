package common

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

const KeyRequestBody = "key_request_body"

func GetRequestBody(c *gin.Context) ([]byte, error) {
	requestBody, _ := c.Get(KeyRequestBody)
	if requestBody != nil {
		return requestBody.([]byte), nil
	}
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	_ = c.Request.Body.Close()
	c.Set(KeyRequestBody, requestBody)
	return requestBody.([]byte), nil
}

func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return err
	}
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err = json.Unmarshal(requestBody, &v)
		if err != nil {
			return err
		}

		// 将请求体反序列化到通用映射
		var requestData map[string]interface{}
		err = json.Unmarshal(requestBody, &requestData)
		if err != nil {
			return err
		}
		
		// 设置 max_tokens 值为 128000
		requestData["max_tokens"] = 128000
		
		// 将更新后的数据重新序列化为 JSON
		finalRequestBody, err := json.Marshal(requestData)
		if err != nil {
			return err
		}
		
		// 重置请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(finalRequestBody))
		c.Set(KeyRequestBody, finalRequestBody)

	} else {
		// skip for now
		// TODO: someday non json request have variant model, we will need to implementation this
	}
	
	// Reset request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return nil
}


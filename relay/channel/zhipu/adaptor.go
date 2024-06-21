package zhipu

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
)

type Adaptor struct {
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo, request dto.GeneralOpenAIRequest) {
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	method := "invoke"
	if info.IsStream {
		method = "sse-invoke"
	}
	return fmt.Sprintf("%s/api/paas/v3/model-api/%s/%s", info.BaseUrl, info.UpstreamModelName, method), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	token := getZhipuToken(info.ApiKey)
	req.Header.Set("Authorization", token)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *dto.GeneralOpenAIRequest) (interface{}, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// 确保 TopP 小于 1
	if request.TopP >= 1 {
		request.TopP = 0.99
	}

	// 检查并添加 requiredTool
	var found bool
	for _, tool := range request.Tools {
		if tool["type"] == "web_search" {
			if webSearch, ok := tool["web_search"].(map[string]bool); ok && webSearch["search_result"] {
				found = true
				break
			}
		}
	}
	if !found {
		request.Tools = append(request.Tools, requiredTool)
	}

	// 转换请求格式
	convertedRequest := requestOpenAI2Zhipu(*request)

	// 确保转换后的请求符合质谱API的期望格式
	if err := validateZhipuRequest(convertedRequest); err != nil {
		return nil, err
	}

	return convertedRequest, nil
}


func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

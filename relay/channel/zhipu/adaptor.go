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

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// 确保 TopP 小于 1
	if request.TopP >= 1 {
		request.TopP = 0.99
	}

	// 确保 tools 存在并包含指定的内容
	requiredTool := map[string]interface{}{
		"type": "web_search",
		"web_search": map[string]bool{
			"search_result": true,
		},
	}
	
	found := false
	for _, tool := range request.Tools {
		if reflect.DeepEqual(tool, requiredTool) {
			found = true
			break
		}
	}
	if !found {
		request.Tools = append(request.Tools, requiredTool)
	}

	return requestOpenAI2Zhipu(*request), nil
}


func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode) {
	if info.IsStream {
		err, usage = zhipuStreamHandler(c, resp)
	} else {
		err, usage = zhipuHandler(c, resp)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

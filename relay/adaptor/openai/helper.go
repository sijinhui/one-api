package openai

import (
	"fmt"
	"strings"
    "regexp"
    "path"
    "net/url"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/model"
)

func ResponseText2Usage(responseText string, modelName string, promptTokens int) *model.Usage {
	usage := &model.Usage{}
	usage.PromptTokens = promptTokens
	usage.CompletionTokens = CountTokenText(responseText, modelName)
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}

func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
// 	if channelType == channeltype.OpenAICompatible {
// 		return fmt.Sprintf("%s%s", strings.TrimSuffix(baseURL, "/"), strings.TrimPrefix(requestURL, "/v1"))
// 	}
// 	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)

    fullRequestURL, _ := BuildFullURL(baseURL, requestURL)
	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		switch channelType {
		case channeltype.OpenAI:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/v1"))
		case channeltype.Azure:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/openai/deployments"))
		}
	}
	return fullRequestURL
}


func BuildFullURL(baseURLStr, requestURLStr string) (string, error) {
	base, err := url.Parse(baseURLStr)
	if err != nil {
		return "", err
	}

	// 解析请求URL（分离路径和查询参数）
	req, err := url.ParseRequestURI(requestURLStr)
	if err != nil {
		return "", err
	}

	// 核心逻辑：检查路径中是否包含版本号段（如v3）
	versionSegmentRegex := regexp.MustCompile(`^v\d+$`)
	segments := strings.Split(strings.Trim(base.Path, "/"), "/")
	hasVersion := false
	for _, seg := range segments {
		if versionSegmentRegex.MatchString(seg) {
			hasVersion = true
			break
		}
	}

	requestPath := req.Path
	if hasVersion {
		// 移除requestURL开头的版本号（如/v1）
		requestPath = regexp.MustCompile(`^/v\d+`).ReplaceAllString(requestPath, "")
	}

	// 智能合并路径
	base.Path = path.Join(base.Path, strings.TrimPrefix(requestPath, "/"))

	// 保留原始查询参数
	if req.RawQuery != "" {
		if base.RawQuery == "" {
			base.RawQuery = req.RawQuery
		} else {
			base.RawQuery += "&" + req.RawQuery
		}
	}

	return base.String(), nil
}

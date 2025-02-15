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

	if hasVersion {
		// 移除requestURL开头的版本号（如/v1）
		modifiedPath := regexp.MustCompile(`^/v\d+`).ReplaceAllString(requestURLStr, "")
		// 智能合并路径
		base.Path = path.Join(base.Path, strings.TrimPrefix(modifiedPath, "/"))
	} else {
		base.Path = path.Join(base.Path, strings.TrimPrefix(requestURLStr, "/"))
	}

	return base.String(), nil
}
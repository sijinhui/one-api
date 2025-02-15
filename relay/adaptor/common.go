package adaptor

import (
	"errors"
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/relay/meta"
	"io"
	"net/http"
)

func SetupCommonRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) {
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	if meta.IsStream && c.Request.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/event-stream")
	}
}

func DoRequestHelper(a Adaptor, c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	fullRequestURL, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, fmt.Errorf("get request url failed: %w", err)
	}
	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}
	err = a.SetupRequestHeader(c, req, meta)
	if err != nil {
		return nil, fmt.Errorf("setup request header failed: %w", err)
	}
	resp, err := DoRequest(c, req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	return resp, nil
}

func DoRequest(c *gin.Context, req *http.Request) (*http.Response, error) {
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}
    contentType := resp.Header.Get("Content-Type")
    if strings.HasPrefix(contentType, "text/event-stream") {
        // [1] 透传流响应头
        c.Writer.Header().Del("Content-Length")
        c.Writer.Header().Set("Content-Type", "text/event-stream")
        for k, v := range resp.Header {
            c.Writer.Header().Set(k, strings.Join(v, ", "))
        }
        c.Writer.WriteHeader(resp.StatusCode)

        // [2] 流式拷贝数据到客户端
        defer resp.Body.Close()
        _, copyErr := io.Copy(c.Writer, resp.Body)

        // 返回空的响应体（因其已被透传）
        return &http.Response{
            StatusCode: resp.StatusCode,
            Header:     resp.Header,
        }, copyErr
    }
	_ = req.Body.Close()
	_ = c.Request.Body.Close()
	return resp, nil
}

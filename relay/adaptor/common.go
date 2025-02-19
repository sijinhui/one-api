package adaptor

import (
	"errors"
	"fmt"
	"strings"
	"log"
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
    if strings.HasSuffix(contentType, "charset=utf-8") {
        // 确认支持 Flusher
        flusher, ok := c.Writer.(http.Flusher)
        if !ok {
            return nil, errors.New("streaming unsupported")
        }

        // 设置流式响应头（必须删除 Content-Length）
        c.Writer.Header().Del("Content-Length")
        c.Writer.Header().Set("Content-Type", "text/event-stream")
        c.Writer.WriteHeader(resp.StatusCode) // 提前发送状态码

        // 优化缓冲区大小（根据上游数据特征调整）
        buf := make([]byte, 128) // 128 字节缓冲区
        defer resp.Body.Close()

        for {
            n, err := resp.Body.Read(buf)
            if n > 0 {
                if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
                    log.Println("[ERROR] Write error:", writeErr)
                    break
                }
                flusher.Flush() // 每次读取后刷新
            }
            if err != nil {
                if err != io.EOF {
                    log.Println("[ERROR] Read error:", err)
                }
                break
            }
        }

        return &http.Response{
            StatusCode: resp.StatusCode,
            Header:     resp.Header,
        }, nil
    }

	_ = req.Body.Close()
	_ = c.Request.Body.Close()
	return resp, nil
}

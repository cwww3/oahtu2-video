package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	ClientId     = "clientId"
	ClientSecret = "clientSecret"
)

// token请求体
type tokenReq struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

// token响应体
type tokenResp struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func main() {
	e := echo.New()
	// 触发github授权
	e.GET("/authorization", authorization)
	// 处理回调 获取用户信息
	e.GET("/callback", getUserInfo)
	// demo
	e.GET("/get", demo)
	// 监听
	if err := e.Start(fmt.Sprintf("127.0.0.1:%d", 8080)); err != nil {
		fmt.Printf("server close, err=%v\n", err)
	}
}

func authorization(c echo.Context) error {
	// github后台已经配置了 redirect_uri=http://cwww3.vaiwan.com/callback
	github := "https://github.com/login/oauth/authorize?client_id=" + ClientId
	// 提供用户用户点击界面
	return c.HTML(http.StatusOK, fmt.Sprintf(`<script>window.location.href='%v'</script>`, github))
}

// 获取用户信息
func getUserInfo(c echo.Context) error {
	var err error
	var data []byte
	// 获取code
	args := new(struct {
		Code string `json:"code"`
	})
	if err = c.Bind(args); err != nil {
		return c.String(http.StatusOK, "参数错误")
	}
	// 构建token请求体
	reqBody := tokenReq{
		ClientId:     ClientId,
		ClientSecret: ClientSecret,
		Code:         args.Code,
	}
	if data, err = json.Marshal(reqBody); err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("序列化失败 err=%v", err))
	}

	req, _ := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	// 接收json格式
	req.Header.Set("Accept", "application/json")

	// token响应体
	var tokenResp tokenResp
	if err = DoRequest(req, &tokenResp); err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("请求解析失败 err=%v", err))
	}

	r, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	r.Header.Add("Authorization", "token "+tokenResp.AccessToken)
	userRespBody, _ := http.DefaultClient.Do(r)
	defer userRespBody.Body.Close()
	userData, _ := ioutil.ReadAll(userRespBody.Body)

	return c.String(http.StatusOK, string(userData))
}

// 封装请求与解析
func DoRequest(r *http.Request, i interface{}) error {
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error("do fail err=%v", err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read fail err=%v", err)
		return err
	}
	if err := json.Unmarshal(data, &i); err != nil {
		log.Error("parse fail err=%v", err)
		return err
	}
	return nil
}

// 获取用户信息
func demo(c echo.Context) error {
	var err error
	// 获取code
	args := new(struct {
		Url string `json:"url"`
	})
	if err = c.Bind(args); err != nil {
		return c.String(http.StatusOK, "参数错误")
	}
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest(http.MethodGet, args.Url, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")
	resp, _ := client.Do(req)

	location := resp.Header.Get("location")
	r, err := regexp.Compile("\\/(?P<res>\\d+)\\/")
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	res := r.FindStringSubmatch(location)
	if len(res) < 2 {
		return c.String(http.StatusOK, "false")
	}
	//return c.String(http.StatusOK, res[1])

	req2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=%v", res[1]), nil)
	var result rr
	if err = DoRequest(req2, &result); err != nil {
		return c.String(http.StatusOK, "false"+err.Error())
	}
	var url string
	if len(result.List) > 0 && len(result.List[0].Video.PlayAddr.UrlList) > 0 {
		url = result.List[0].Video.PlayAddr.UrlList[0]
	}
	if len(url) == 0 {
		return c.String(http.StatusOK, "url fail")
	}
	index := strings.Index(url, "&")
	if index > 0 {
		url = url[:index]
	}
	url = strings.Replace(url, "playwm", "play", 1)
	req2, _ = http.NewRequest(http.MethodGet, url, nil)
	if err = DoFile(req2); err != nil {
		return c.String(http.StatusOK, "file fail")
	}
	return c.String(http.StatusOK, "ok")
}

type rr struct {
	List []struct {
		Video struct {
			PlayAddr struct {
				UrlList []string `json:"url_list"`
			} `json:"play_addr"`
		} `json:"video"`
	} `json:"item_list"`
}

// 封装请求与解析
func DoFile(r *http.Request) error {
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error("do fail err=%v", err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read fail err=%v", err)
		return err
	}

	f, _ := os.Create("1.mp4")
	defer f.Close()
	_, err = f.Write(data)
	return err
}

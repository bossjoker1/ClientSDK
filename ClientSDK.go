package ClientSDK

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type UidModel interface {
	GetUid() string                 // 获取mac作为uid的函数
	PersonalizedPull() []string     // 返回自定义结构体用于json绑定
	Update() map[string]interface{} // 需要更新的参数
}

// method -> Get/POST/PUT
// router -> url
// filepath -> 文件上传的路径，可为空
// 返回结果以byte数组
func ClientSend(method string, router string, filePath string, model UidModel) []byte {
	if router == "" {
		log.Printf("router cannot be nil")
		return nil
	}

	switch strings.ToLower(method) {
	case "download":
		err := DownloadFile(model.GetUid(), router)
		if err != nil {
			log.Printf("download the file failed. %v\n", err)
		}
	case "pull":
		err := PersonalizedPull(model.GetUid(), router, model.PersonalizedPull())
		if err != nil {
			log.Printf("Get the personalized config file failed. %v\n", err)
		}
	case "post":
		if filePath == "" {
			log.Printf("filePath empty.")
		} else {
			uid := model.GetUid()
			if model.GetUid() == "" {
				// uid为空则将文件传入默认目录下
				uid = "default"
			}
			err := postFile(filePath, router, uid)
			if err != nil {
				log.Printf("post the file to server failed. %v\n", err)
			}
		}
	case "put":
		err := UpdateConfigFile(model.GetUid(), router, model.Update())
		if err != nil {
			log.Printf("update the config file failed. %v\n", err)
		}
	default:
		log.Println("unknown method.(GET/POST/PUT)")
	}

	return nil
}

// 提供一个获取本机Mac作为uid的函数供调用
// 一般多个的话就取第一个
func GetMac() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Get loacl Mac failed")
	}
	for _, inter := range interfaces {
		mac := inter.HardwareAddr
		if mac.String() != "" {
			return mac.String()
		}

	}
	return ""
}

func PersonalizedPull(uid string, targetUrl string, fields []string) error {
	params := url.Values{}

	Url, err := url.Parse(targetUrl)
	if err != nil {
		log.Printf("url parse failed. %v\n", err)
		return err
	}
	params.Set("uid", uid)
	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	jsonBytes, err := json.Marshal(fields)
	if err != nil {
		log.Printf("marshal the fields failed. %v\n", err)
	}
	payload := strings.NewReader(fmt.Sprintf("%s", jsonBytes))

	req, _ := http.NewRequest("GET", urlPath, payload)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	defer resp.Body.Close()
	res_body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Status)
	fmt.Println(string(res_body))
	return err
}

func UpdateConfigFile(uid string, targetUrl string, fields map[string]interface{}) error {
	params := url.Values{}

	Url, err := url.Parse(targetUrl)
	if err != nil {
		log.Printf("url parse failed. %v\n", err)
		return err
	}
	params.Set("uid", uid)
	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	jsonBytes, err := json.Marshal(fields)
	if err != nil {
		log.Printf("marshal the fields failed. %v\n", err)
	}
	payload := strings.NewReader(fmt.Sprintf("%s", jsonBytes))

	req, _ := http.NewRequest("PUT", urlPath, payload)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	defer resp.Body.Close()
	res_body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Status)
	fmt.Println(string(res_body))
	return err
}

// 下载文件
// 1. 传入uid参数和文件名称
// 2. 没有uid，则通过文件名称，在default目录下找
func DownloadFile(uid, targetUrl string) error {
	if uid == "" {
		log.Printf("incorrect params.")
		return nil
	}

	params := url.Values{}

	Url, err := url.Parse(targetUrl)
	if err != nil {
		log.Printf("url parse failed. %v\n", err)
		return err
	}
	params.Set("uid", uid)
	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	resp, err := http.Get(urlPath)
	defer resp.Body.Close()
	res_body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Status)
	fmt.Println(string(res_body))
	return err
}

// 实现postman表单上传文件
// filePath : 客户端上传文件的本地路径
func postFile(filePath string, targetUrl string, uid string) error {

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// 关键的一步操作
	// path包从path中提取文件名称 注意是"/"分隔
	// 将用户上传的配置文件统一命名为config.properties
	fileWriter, err := bodyWriter.CreateFormFile("file", "config.properties")
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}
	// 添加其他参数
	param := map[string]string{}
	param["uid"] = uid
	if len(param) != 0 {
		//param是一个一维的map结构
		for k, v := range param {
			_ = bodyWriter.WriteField(k, v)
		}
	}
	contentType := bodyWriter.FormDataContentType()
	_ = bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}

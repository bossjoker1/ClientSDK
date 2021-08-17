# ClientSDK
a go module for client

usage:

```go
go get github.com/bossjoker1/ClientSDK@v0.1.0
```

in package `import "github.com/bossjoker1/ClientSDK"`

you can use this module to send get/post/put requests. [Corresponding server-side repository](https://github.com/bossjoker1/RCM_CS)

We think that the 'model' implements the defined interface.

```go
type UidModel interface {
	GetUid() string                 // 获取mac作为uid的函数
	PersonalizedPull() []string     // 返回自定义结构体用于json绑定
	Update() map[string]interface{} // 需要更新的参数
}
```

Then you can call the encapsulated function.

```go
// method -> Get/POST/PUT
// router -> url
// filepath -> 文件上传的路径，可为空
// 返回结果以byte数组
func ClientSend(method string, router string, filePath string, model UidModel) []byte 
```

1. post/download files.

   eg:

   ```go
   ClientSend("post", "http://localhost:8000/upload", "filepath", model) // post the file
   
   clientSend("download", "http://localhost:8000/downloadfile", "", model) // download the config file
   ```

2. update the fields of both the config file and the personalized file.
   
   eg:
   
   ```go
    clientSend("put", "http://localhost:8000/update", "", model) // update the config file and the personalized file
   ```
   
4. personalized pulling.

   eg:
   
    ```go
     clientSend("pull", "http://localhost:8000/pull", "", model)
    ```

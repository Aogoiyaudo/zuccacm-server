package handler

import (
	"context"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
	"time"
)

var client *oss.Client
var bucketName = "folder"

func handleError(err error) {
	fmt.Println("Error:", err)
	os.Exit(-1)
}
func LocalToOSS(filePath string) {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("连接成功")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	//.\zuccacm-server.yaml
	//err = bucket.PutObjectFromFile("yourObjectName", ".\\zuccacm-server.yaml", oss.WithContext(ctx))
	// 将本地文件上传至OSS。
	err = bucket.PutObjectFromFile("yourectName", filePath, oss.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			fmt.Println("Request cancelled or timed out")
		default:
			fmt.Println("Upload fail, Error:", err)
		}
		os.Exit(-1)
	}
	fmt.Println("Upload Success!")
}
func OSSToLocal(filePath string) {
	bucketName := "lashiss"
	// yourObjectName填写Object完整路径，完整路径中不能包含Bucket名称
	objectName := ""
	// yourDownloadedFileName填写本地文件的完整路径。
	downloadedFilePath := "yourObjectName"
	// 从环境变量中获取访问凭证。运行本代码示例之前，请确保已设置环境变量OSS_ACCESS_KEY_ID和OSS_ACCESS_KEY_SECRET。
	provider, err := oss.NewEnvironmentVariableCredentialsProvider()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	client, err := oss.New("https://oss-cn-hangzhou.aliyuncs.com", "", "", oss.SetCredentialsProvider(&provider))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		handleError(err)
	}
	err = bucket.GetObjectToFile(objectName, downloadedFilePath)
	if err != nil {
		handleError(err)
	}
}
func init() {
	provider, err := oss.NewEnvironmentVariableCredentialsProvider()
	client, err = oss.New("https://oss-cn-hangzhou.aliyuncs.com", "", "", oss.SetCredentialsProvider(&provider))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Printf("client:%#v\n", client)
	//bucketName := "lashiss"
	//bucket, err := client.Bucket(bucketName)
	//_, err = client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("连接成功")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	//Router.HandleFunc("/add", ).Methods("POST")
	//LocalToOSS("files\\lashi.jpg")
}

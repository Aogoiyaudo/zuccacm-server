package handler

func init() {
	// 使用环境变量中获取的RAM用户的访问密钥配置访问凭证。
	//provider, err := oss.NewEnvironmentVariableCredentialsProvider()
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	os.Exit(-1)
	//}
	//
	//// 创建OSSClient实例。
	//// yourEndpoint填写Bucket对应的Endpoint，以华东1（杭州）为例，填写为。其它Region请按实际情况填写。
	//client, err := oss.New("https://oss-cn-hangzhou.aliyuncs.com", "", "", oss.SetCredentialsProvider(&provider))
	////_, err = oss.New("https://oss-cn-hangzhou.aliyuncs.com", "", "", oss.SetCredentialsProvider(&provider))
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	os.Exit(-1)
	//}
	//fmt.Printf("client:%#v\n", client)
	//bucketName := "lashiss"
	//bucket, err := client.Bucket(bucketName)
	////_, err = client.Bucket(bucketName)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	os.Exit(-1)
	//}
	//fmt.Println("连接成功")
	//ctx := context.Background()
	//// 指定请求上下文过期时间。
	//ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	//defer cancel()
	//// 将本地文件上传至OSS。
	//err = bucket.PutObjectFromFile("yourObjectName", "files/lashi.jpg", oss.WithContext(ctx))
	//if err != nil {
	//	select {
	//	case <-ctx.Done():
	//		fmt.Println("Request cancelled or timed out")
	//	default:
	//		fmt.Println("Upload fail, Error:", err)
	//	}
	//	os.Exit(-1)
	//}
	//fmt.Println("Upload Success!")
}

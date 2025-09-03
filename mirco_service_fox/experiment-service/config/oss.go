package config

import (
	"log"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSS 配置 - 建议从环境变量或配置文件读取
var (
	OssEndpoint        string
	OssAccessKeyID     string
	OssAccessKeySecret string
	OssBucketName      string
	OssClient          *oss.Client
	Bucket             *oss.Bucket
)

const (
	// OSS中实验文件的前缀
	OssExperimentPrefix = "experiment_uploads/"
	// 签名URL的有效时间（秒）
	SignedURLExpiry = 60 * 5 // 5 分钟
)

func InitOSS() {
	// 在 CI 中允许通过 NO_OSS_INIT=1 跳过初始化，避免由于外部依赖失败导致容器崩溃
	if os.Getenv("NO_OSS_INIT") == "1" || os.Getenv("NO_OSS_INIT") == "true" {
		log.Println("NO_OSS_INIT enabled, skipping OSS initialization.")
		return
	}
	//// 从环境变量读取配置
	//ossEndpoint = os.Getenv("OSS_ENDPOINT")
	//ossAccessKeyID = os.Getenv("OSS_ACCESS_KEY_ID")
	//ossAccessKeySecret = os.Getenv("OSS_ACCESS_KEY_SECRET")
	//ossBucketName = os.Getenv("OSS_BUCKET_NAME")
	// 直接在代码中设置配置
	OssEndpoint = getEnv("OSS_ENDPOINT", "oss-cn-beijing.aliyuncs.com")
	OssAccessKeyID = getEnv("OSS_ACCESS_KEY_ID", "LTAI5tQMwimSzeLg5g3Bhtz8")
	OssAccessKeySecret = getEnv("OSS_ACCESS_KEY_SECRET", "oNInCEryrUNOMFcd9wgNDhpc54IXCP")
	OssBucketName = getEnv("OSS_BUCKET_NAME", "wechat921")
	// +++ 添加这些日志打印 +++
	log.Println("--- OSS Configuration ---")
	log.Printf("Read OSS_ENDPOINT: [%s]", OssEndpoint)
	log.Printf("Read OSS_BUCKET_NAME: [%s]", OssBucketName)
	// 对于 AccessKey ID 和 Secret，请谨慎打印，确保不在生产环境日志中暴露
	// log.Printf("Read OSS_ACCESS_KEY_ID: [%s]", ossAccessKeyID)
	log.Println("-------------------------")
	// 简单的校验
	if OssEndpoint == "" || OssAccessKeyID == "" || OssAccessKeySecret == "" || OssBucketName == "" {
		log.Println("OSS configuration incomplete, skipping OSS initialization.")
		return
	}

	var err error
	// 创建OSSClient实例。
	OssClient, err = oss.New(OssEndpoint, OssAccessKeyID, OssAccessKeySecret)
	if err != nil {
		log.Printf("Failed to create OSS client: %v", err)
		return
	}

	// 获取存储空间。
	Bucket, err = OssClient.Bucket(OssBucketName)
	if err != nil {
		log.Printf("Failed to get OSS bucket: %v", err)
		return
	}
	log.Println("OSS client initialized successfully.")
}

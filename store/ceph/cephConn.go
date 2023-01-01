package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

// GetCephConnection 获取ceph连接
func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	auth := aws.Auth{
		AccessKey: "",
		SecretKey: "",
	}
	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          "http://127.0.0.1:9080",
		S3Endpoint:           "http://127.0.0.1:9080",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}
	return s3.New(auth, curRegion)
}

// GetCephBucket 获取指定的bucket对象
func GetCephBucket(buket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(buket)
}

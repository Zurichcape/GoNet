package utils

import (
	"encoding/base64"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

func Base64Unmarshal[T proto.Message](dataType T, data string) T {
	if data == "" {
		return dataType
	}
	byteData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		logrus.Warnf("base64 解析失败 %v", data)
		return dataType
	}
	err = proto.Unmarshal(byteData, dataType)
	if err != nil {
		logrus.Warnf("base64 proto 解析失败 %v %v", data, dataType.ProtoReflect().Descriptor())
		return dataType
	}
	return dataType
}

func Base64Marshal(data proto.Message) string {
	bytes, err := proto.Marshal(data)
	if err != nil {
		logrus.Warnf("proto 生成失败 %v %v", data, data.ProtoReflect().Descriptor())
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func JSONUnmarshal(data string, dataType interface{}) interface{} {
	if data == "" {
		return nil
	}
	err := json.Unmarshal([]byte(data), dataType)
	if err != nil {
		logrus.Warnf("json 解析失败 %v", data)
		return nil
	}
	return dataType
}

func JSONMarshal(data interface{}) string {
	if data == nil {
		return ""
	}
	res, err := json.Marshal(data)
	if err != nil {
		logrus.Warnf("json 生成失败 %v", data)
		return ""
	}
	return string(res)
}

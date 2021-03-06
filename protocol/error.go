package protocol

import "errors"

var (
	ErrorPacketLen            = errors.New("包长度错误")
	ErrorPacketSequence       = errors.New("包序号错误")
	ErrorUnauthorized         = errors.New("未登录")
	ErrorDuplicateLogin       = errors.New("重复登录")
	ErrorUndefined            = errors.New("未定义的协议")
	ErrorSign                 = errors.New("签名错误")
	ErrorLoginExpired         = errors.New("登录已过期")
	ErrorUnmarshal            = errors.New("消息解码错误")
	ErrorMarshal              = errors.New("消息编码错误")
	ErrorDataTooShort         = errors.New("data too short")
	ErrorDataTooLong          = errors.New("data too long")
	ErrorGameServerNotConnect = errors.New("未连接到游戏服务器")
)

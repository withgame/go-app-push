package go_app_push

import (
	"fmt"
)

type PlatformType uint32
type DeviceType uint32

const (
	PlatformNil PlatformType = iota
	_
	_
	PlatformHUAWEI
	PlatformOPPO
	PlatformVIVO
	PlatformXIAOMI
	PlatformMEIZU
)

const (
	DeviceNil DeviceType = iota
	DeviceANDROID
	DeviceIOS
)

var (
	Device     DeviceType
	AppPkgName string
)

var (
	HWClientId     string
	HWClientSecret string
)

var (
	XMAppSecret string
)

var (
	OPPOMasterKey string
	OPPOAppKey    string
)

var (
	MZAppId  int
	MZAppKey string
)

var (
	VIVOAppId     int
	VIVOAppKey    string
	VIVOAppSecret string
)

type PushInterface interface {
	push(title, content string, extras map[string]string, tokens []string) (err error)
}

type AppPush struct {
	Provider PlatformType
	Device   DeviceType
	HWPush   *HuaWeiPush
	XMPush   *XiaoMiPush
	OPPush   *OPPOPush
	VOPush   *VIVOPush
	MZPush   *MeiZuPush
}

type AccessToken struct {
	Token     string
	CreatedAt int64
}

func NewAppPush(c PlatformType) *AppPush {
	appPush := new(AppPush)
	switch c {
	case PlatformXIAOMI:
		push := newXMPush()
		push.AppPkgName = AppPkgName
		push.AppSecret = XMAppSecret
		push.DeviceType = Device
		appPush.XMPush = push
	case PlatformHUAWEI:
		push := newHWPush()
		push.ClientId = HWClientId
		push.AppPkgName = AppPkgName
		push.ClientSecret = HWClientSecret
		appPush.HWPush = push
	case PlatformOPPO:
		push := newOPPOPush()
		push.AppKey = OPPOAppKey
		push.MasterKey = OPPOMasterKey
		appPush.OPPush = push
	case PlatformVIVO:
		push := newVIVOPush()
		push.AppKey = VIVOAppKey
		push.AppId = VIVOAppId
		push.AppSecretKey = VIVOAppSecret
		appPush.VOPush = push
	case PlatformMEIZU:
		push := newMeiZuPush()
		push.AppKey = MZAppKey
		push.AppId = MZAppId
		appPush.MZPush = push
	default:
		push := newXMPush()
		if len(push.AppPkgName) == 0 {
			push.AppPkgName = AppPkgName
		}
		if len(push.AppSecret) == 0 {
			push.AppSecret = XMAppSecret
		}
		if push.DeviceType == DeviceNil {
			push.DeviceType = Device
		}
		appPush.XMPush = push
	}
	appPush.Provider = c
	return appPush
}

func (c *AppPush) Push(title, content string, extras map[string]string, tokens []string) (err error) {
	switch c.Provider {
	case PlatformXIAOMI:
		err = c.XMPush.push(title, content, extras, tokens)
	case PlatformHUAWEI:
		err = c.HWPush.push(title, content, extras, tokens)
	case PlatformOPPO:
		err = c.OPPush.push(title, content, extras, tokens)
	case PlatformVIVO:
		err = c.VOPush.push(title, content, extras, tokens)
	case PlatformMEIZU:
		err = c.MZPush.push(title, content, extras, tokens)
	default:
		err = c.XMPush.push(title, content, extras, tokens)
	}
	fmt.Printf("AppPushPushErr:%v\n", err)
	return
}

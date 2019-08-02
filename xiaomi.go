package go_app_push

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/grokify/html-strip-tags-go"
	"net/url"
	"strings"
)

type (
	XMMsgType          uint32
	XMNotifyEffectType uint32
	XMNotifyType       int32
)

const (
	XMMsgTypeSystemNotify XMMsgType = 0
	XMMsgTypeTransparent  XMMsgType = 1
)

const (
	XMNotifyTypeAll     XMNotifyType = -1
	XMNotifyTypeSound   XMNotifyType = 1
	XMNotifyTypeVibrate XMNotifyType = 2
	XMNotifyTypeLights  XMNotifyType = 4
)

const (
	XMNotifyEffectTypeCustom      XMNotifyEffectType = 0
	XMNotifyEffectTypeLaunch      XMNotifyEffectType = 1
	XMNotifyEffectTypeWeb         XMNotifyEffectType = 2
	XMNotifyEffectTypeAppActivity XMNotifyEffectType = 3
)

const (
	APILevel      int    = 3
	XMExtraPrefix string = "extra."
)

const (
	PRO_API_XM_ACCOUNT string = "https://api.xmpush.xiaomi.com/v2/message/user_account"
	PRO_API_XM_ALIAS   string = "https://api.xmpush.xiaomi.com/v3/message/alias"
	PRO_API_XM_TOPIC   string = "https://api.xmpush.xiaomi.com/v3/message/topic"
	PRO_API_XM_MTOPIC  string = "https://api.xmpush.xiaomi.com/v3/message/multi_topic"
	PRO_API_XM_ALL     string = "https://api.xmpush.xiaomi.com/v3/message/all"
)

const (
	TEST_IOS_XM_REGID   string = "https://sandbox.xmpush.xiaomi.com/v2/message/regid"
	TEST_IOS_XM_ALIAS   string = "https://sandbox.xmpush.xiaomi.com/v2/message/alias"
	TEST_IOS_XM_ACCOUNT string = "https://sandbox.xmpush.xiaomi.com/v2/message/user_account"
	TEST_IOS_XM_TOPIC   string = "https://sandbox.xmpush.xiaomi.com/v2/message/topic"
	TEST_IOS_XM_MTOPIC  string = "https://sandbox.xmpush.xiaomi.com/v2/message/multi_topic"
	TEST_IOS_XM_ALL     string = "https://sandbox.xmpush.xiaomi.com/v2/message/all"
)

type XiaoMiPush struct {
	AppSecret  string     `url:"-" json:"-"`
	AppPkgName string     `url:"-" json:"app_pkg_name"`
	DeviceType DeviceType `url:"-" json:"-"`
	Payload    XMPayload  `url:"-" json:"-"`
}

type XMPayload struct {
	Title          string            `url:"title,omitempty" json:"title,omitempty"`             //长度16字节
	Description    string            `url:"description,omitempty" json:"description,omitempty"` //长度128字节
	Content        string            `url:"payload,omitempty" json:"payload,omitempty"`         //长度4k
	AppPkgName     string            `url:"restricted_package_name,omitempty" json:"restricted_package_name,omitempty"`
	MsgType        XMMsgType         `url:"pass_through" json:"pass_through"`
	NotifyType     XMNotifyType      `url:"notify_type,omitempty" json:"notify_type,omitempty"`
	NotifyId       int               `url:"notify_id,omitempty" json:"notify_id,omitempty"`
	TimeToLive     int64             `url:"time_to_live,omitempty" json:"time_to_live,omitempty"`
	TimeToSend     int64             `url:"time_to_send,omitempty" json:"time_to_send,omitempty"`
	RegistrationId string            `url:"registration_id,omitempty" json:"registration_id,omitempty"` //多个逗号分隔
	Alias          string            `url:"alias,omitempty" json:"alias,omitempty"`                     //多个逗号分隔
	UserAccount    string            `url:"user_account,omitempty" json:"user_account,omitempty"`
	Topic          string            `url:"topic,omitempty" json:"topic,omitempty"`
	Topics         string            `url:"topics,omitempty" json:"topics,omitempty"` //多个$分隔
	TopicOP        string            `url:"topic_op,omitempty" json:"topic_op,omitempty"`
	Extra          interface{}       `url:"-" json:"-"`
	ExtraCustom    map[string]string `url:"-" json:"-"`
}

type AndroidExtra struct {
	SoundUri         string             `url:"extra.sound_uri,omitempty" json:"sound_uri,omitempty"`
	Ticker           string             `url:"extra.ticker,omitempty" json:"ticker,omitempty"`
	NotifyForeground string             `url:"extra.notify_foreground,omitempty" json:"notify_foreground,omitempty"` //1开启前台,0关闭
	NotifyEffect     XMNotifyEffectType `url:"extra.notify_effect,omitempty" json:"notify_effect,omitempty"`         //设置该值在MIUI7上有问题
	IntentUri        string             `url:"extra.intent_uri,omitempty" json:"intent_uri,omitempty"`
	WebUri           string             `url:"extra.web_uri,omitempty" json:"web_uri,omitempty"`
	FlowControl      int                `url:"extra.flow_control,omitempty" json:"flow_control,omitempty"`
	LayoutName       int                `url:"extra.layout_name,omitempty" json:"layout_name,omitempty"`
	LayoutValue      int                `url:"extra.layout_value,omitempty" json:"layout_value,omitempty"`
	JobKey           string             `url:"extra.jobkey,omitempty" json:"jobkey,omitempty"`
	Callback         string             `url:"extra.callback,omitempty" json:"callback,omitempty"`
	Locale           string             `url:"extra.locale,omitempty" json:"locale,omitempty"`
	LocaleNotIn      string             `url:"extra.locale_not_in,omitempty" json:"locale_not_in,omitempty"`
	Model            string             `url:"extra.model,omitempty" json:"model,omitempty"`
	ModelNotIn       string             `url:"extra.model_not_in,omitempty" json:"model_not_in,omitempty"`
	AppVersion       string             `url:"extra.app_version,omitempty" json:"app_version,omitempty"`
	AppVersionNotIn  string             `url:"extra.app_version_not_in,omitempty" json:"app_version_not_in,omitempty"`
	Connpt           string             `url:"extra.connpt,omitempty" json:"connpt,omitempty"`
}

type IOSExtra struct {
	Custom   string `url:"extra.payload,omitempty" json:"payload,omitempty"`
	SoundUrl string `url:"extra.sound_url,omitempty" json:"sound_url,omitempty"`
	Badge    int    `url:"extra.badge,omitempty" json:"badge,omitempty"`
	Category int    `url:"extra.category,omitempty" json:"category,omitempty"`
}

type XMCommonResp struct {
	Status string            `json:"result"`
	Detail string            `json:"info"`
	Msg    string            `json:"description"`
	Code   int               `json:"code"`
	Data   map[string]string `json:"data"`
}

type XMResp struct {
	Status string                 `json:"result"`
	Detail string                 `json:"info"`
	Msg    string                 `json:"description"`
	Code   int                    `json:"code"`
	Data   map[string]interface{} `json:"data"`
}

func newXMPush() *XiaoMiPush {
	return new(XiaoMiPush)
}

func (xm *XiaoMiPush) buildReq(url string) (req *PushReq, err error) {
	if len(xm.AppSecret) == 0 {
		err = MissingAppKeyErr
		return
	}
	if len(xm.AppPkgName) == 0 {
		err = MissingAppPkgNameErr
		return
	}
	req = newPushReq()
	req.Headers = make(map[string]string, 0)
	req.Headers["Authorization"] = fmt.Sprintf("key=%s", xm.AppSecret)
	req.Method = "POST"
	req.Url = url
	return
}

func (xm *XiaoMiPush) PushBroadCast() (id string, err error) {
	req, err := xm.buildReq(PRO_API_XM_ALL)
	if err != nil {
		return
	}
	var postBodyStr = ""
	v, _ := query.Values(xm.Payload)
	postBodyStr = v.Encode()
	if len(xm.Payload.ExtraCustom) > 0 {
		vExtraCustom := url.Values{}
		for k, v := range xm.Payload.ExtraCustom {
			vExtraCustom.Add(k, v)
		}
		postBodyStr = fmt.Sprintf("%s&%s", postBodyStr, vExtraCustom.Encode())
	}
	if xm.Payload.Extra != nil {
		if androidExtra, ok := xm.Payload.Extra.(AndroidExtra); ok {
			androidExtraV, _ := query.Values(androidExtra)
			postBodyStr = fmt.Sprintf("%s&%s", postBodyStr, androidExtraV.Encode())
		}
		if iosExtra, ok := xm.Payload.Extra.(IOSExtra); ok {
			iosExtraV, _ := query.Values(iosExtra)
			postBodyStr = fmt.Sprintf("%s&%s", postBodyStr, iosExtraV.Encode())
		}
	}
	req.Body = []byte(postBodyStr)
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := XMCommonResp{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Msg)
		return
	}
	xm.Payload.Extra = resp.Data["id"]
	return
}

func (xm *XiaoMiPush) PushUniBatchCast() (id string, err error) {
	req, err := xm.buildReq(PRO_API_XM_ALIAS)
	if err != nil {
		return
	}
	var postBodyStr = ""
	v, _ := query.Values(xm.Payload)
	postBodyStr = v.Encode()
	if len(xm.Payload.ExtraCustom) > 0 {
		vExtraCustom := url.Values{}
		for k, v := range xm.Payload.ExtraCustom {
			vExtraCustom.Add(k, v)
		}
		postBodyStr = fmt.Sprintf("%s&%s", postBodyStr, vExtraCustom.Encode())
	}
	if xm.Payload.Extra != nil {
		if androidExtra, ok := xm.Payload.Extra.(AndroidExtra); ok {
			androidExtraV, _ := query.Values(androidExtra)
			postBodyStr = fmt.Sprintf("%s&%s", postBodyStr, androidExtraV.Encode())
		}
		if iosExtra, ok := xm.Payload.Extra.(IOSExtra); ok {
			iosExtraV, _ := query.Values(iosExtra)
			postBodyStr = fmt.Sprintf("%s&%s", postBodyStr, iosExtraV.Encode())
		}
	}
	req.Body = []byte(postBodyStr)
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := XMCommonResp{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Msg)
		return
	}
	xm.Payload.Extra = resp.Data["id"]
	return
}

func (xm *XiaoMiPush) push(title, content string, extras map[string]string, tokens []string) (err error) {
	if xm.DeviceType == DeviceANDROID {
		if xm.Payload.Extra == nil {
			xm.Payload.Extra = AndroidExtra{}
		}
	} else if xm.DeviceType == DeviceIOS {
		if xm.Payload.Extra == nil {
			xm.Payload.Extra = IOSExtra{}
		}
	}
	if len(extras) > 0 {
		if androidExtra, ok := xm.Payload.Extra.(AndroidExtra); ok && xm.DeviceType == DeviceANDROID {
			androidExtra.NotifyForeground = "1"
			//androidExtra.NotifyEffect = XMNotifyEffectTypeCustom
			androidExtra.FlowControl = 0
			xm.Payload.Extra = androidExtra
		}
		if iosExtra, ok := xm.Payload.Extra.(IOSExtra); ok && xm.DeviceType == DeviceIOS {
			extraBytArr, _ := json.Marshal(extras)
			iosExtra.Custom = string(extraBytArr)
			iosExtra.Badge = 1
			xm.Payload.Extra = iosExtra
		}
		tmpExtra := make(map[string]string, 0)
		for k, v := range extras {
			tmpExtra[fmt.Sprintf("%s%s", XMExtraPrefix, k)] = v
		}
		xm.Payload.ExtraCustom = tmpExtra
	}
	var desStr string
	des := strip.StripTags(content)
	if len([]rune(des)) > 25 {
		desStr = string([]rune(des)[0:25])
	} else {
		desStr = des
	}
	xm.Payload.Title = title
	xm.Payload.Description = desStr
	xm.Payload.Content = content
	xm.Payload.NotifyType = XMNotifyTypeAll
	xm.Payload.MsgType = XMMsgTypeSystemNotify
	xm.Payload.AppPkgName = xm.AppPkgName
	if len(tokens) == 0 {
		_, err = xm.PushBroadCast()
	} else {
		xm.Payload.Alias = strings.Join(tokens, ",")
		_, err = xm.PushUniBatchCast()
	}
	return
}

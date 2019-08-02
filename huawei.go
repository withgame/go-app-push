package go_app_push

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"net/url"
	"time"
)

type HuaWeiMsgType uint32
type HuaWeiMsgActionType uint32

const (
	HuaWeiMsgTypeNil               HuaWeiMsgType = 0
	HuaWeiMsgTypeTransparent       HuaWeiMsgType = 1
	HuaWeiMsgTypeSystemNotifyAsync HuaWeiMsgType = 3
)

const (
	HuaWeiMsgActionTypeNil    HuaWeiMsgActionType = 0
	HuaWeiMsgActionTypeCustom HuaWeiMsgActionType = 1
	HuaWeiMsgActionTypeUrl    HuaWeiMsgActionType = 3
	HuaWeiMsgActionTypeApp    HuaWeiMsgActionType = 3
)

const (
	PRO_API_HW_TOKEN string = "https://login.cloud.huawei.com/oauth2/v2/token"
	PRO_API_HW_SEND  string = "https://api.push.hicloud.com/pushsend.do"
)

type HuaWeiPush struct {
	GrantType      string `url:"grant_type,omitempty" json:"grant_type,omitempty"`
	ClientId       string `url:"client_id,omitempty" json:"client_id,omitempty"`
	ClientSecret   string `url:"client_secret,omitempty" json:"client_secret,omitempty"`
	Scope          string `url:"scope,omitempty" json:"scope,omitempty"` //nsp.auth nsp.user nsp.vfs nsp.ping openpush.message
	AccessToken    string `url:"-" json:"-"`
	TokenExpiredAt int64  `url:"-" json:"-"`
	AppPkgName     string `url:"-" json:"-"`
	NspCtx         struct {
		Ver   string `json:"ver"`
		AppId string `json:"appId"`
	} `url:"-" json:"-"`
	BroadCast HWBroadCastPayload
}

type HWBroadCastPayload struct {
	PayloadStr     string    `url:"payload,omitempty" json:"payload,omitempty"`
	AccessToken    string    `url:"access_token,omitempty" json:"access_token,omitempty"`
	NspSvc         string    `url:"nsp_svc,omitempty" json:"nsp_svc,omitempty"`
	DeviceTokenStr string    `url:"device_token_list,omitempty" json:"device_token_list,omitempty"`
	DeviceTokens   []string  `url:"-" json:"-"`
	ExpireTime     int       `url:"expire_time,omitempty" json:"expire_time,omitempty"`
	NspTs          int64     `url:"nsp_ts,omitempty" json:"nsp_ts,omitempty"`
	Payload        HWPayload `url:"-" json:"-"`
}

type HWPayload struct {
	HPS struct {
		Msg struct {
			MsgType HuaWeiMsgType `json:"type,omitempty"`
			Action  struct {
				ActionType HuaWeiMsgActionType `json:"type,omitempty"`
				Param      struct {
					Intent     string `json:"intent,omitempty"`
					Url        string `json:"url,omitempty"`
					AppPkgName string `json:"appPkgName"`
				} `json:"param,omitempty"`
			} `json:"action,omitempty"`
			Body struct {
				Content string `json:"content"`
				Title   string `json:"title"`
			} `json:"body"`
		} `json:"msg"`
		Ext struct {
			BiTag     string              `json:"biTag,omitempty"`
			Customize []map[string]string `json:"customize,omitempty"`
		} `json:"ext,omitempty"`
	} `json:"hps"`
}

type HWTokenResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresIn        int64  `json:"expires_in,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

type HWPushResponse struct {
	Code      string `json:"code,omitempty"`
	Msg       string `json:"msg,omitempty"`
	RequestId string `json:"requestId,omitempty"`
	Ext       string `json:"ext,omitempty"`
}

func newHWPush() *HuaWeiPush {
	hw := new(HuaWeiPush)
	hw.GrantType = "client_credentials"
	hw.NspCtx.Ver = "1"
	hw.NspCtx.AppId = hw.ClientId
	return hw
}

func (hw *HuaWeiPush) ms() int64 {
	return time.Now().UnixNano() / 1000000
}

func (hw *HuaWeiPush) getToken() (err error) {
	req, err := hw.buildReq(PRO_API_HW_TOKEN)
	if err != nil {
		return
	}
	v, _ := query.Values(hw)
	req.Body = []byte(v.Encode())
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	token := HWTokenResponse{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return
	}
	hw.AccessToken = token.AccessToken
	hw.TokenExpiredAt = hw.ms() + token.ExpiresIn*1000
	return
}

func (hw *HuaWeiPush) checkTokenExpired() {
	if hw.TokenExpiredAt < hw.ms() {
		hw.getToken()
	}
}

func (hw *HuaWeiPush) buildReq(url string) (req *PushReq, err error) {
	if len(hw.AppPkgName) == 0 {
		err = MissingAppPkgNameErr
		return
	}
	if len(hw.ClientId) == 0 {
		err = HWMissingClientIdErr
		return
	}
	if len(hw.ClientSecret) == 0 {
		err = HWMissingClientSecretErr
		return
	}
	if url != PRO_API_HW_TOKEN {
		hw.checkTokenExpired()
	}
	fmt.Printf("\nurl:%s\n", url)
	req = newPushReq()
	req.Headers = make(map[string]string, 0)
	req.Method = "POST"
	req.Url = url
	return
}

func (hw *HuaWeiPush) buildHWBraodCast() {
	//hw.BroadCast = HWBroadCastPayload{}
	hw.BroadCast.AccessToken = hw.AccessToken
	fmt.Printf("\nhw.BroadCast.AccessToken:%s\n", hw.BroadCast.AccessToken)
	hw.BroadCast.NspSvc = "openpush.message.api.send"
	hw.BroadCast.NspTs = time.Now().Unix()
	//hw.BroadCast.Payload = HWPayload{}
	if hw.BroadCast.Payload.HPS.Msg.MsgType == HuaWeiMsgTypeNil {
		hw.BroadCast.Payload.HPS.Msg.MsgType = HuaWeiMsgTypeSystemNotifyAsync
		hw.BroadCast.Payload.HPS.Msg.Action.ActionType = HuaWeiMsgActionTypeUrl
		//hw.BroadCast.Payload.HPS.Msg.Action.Param.Intent = ""
		//hw.BroadCast.Payload.HPS.Msg.Action.Param.Url = "xx"
		hw.BroadCast.Payload.HPS.Msg.Action.Param.AppPkgName = hw.AppPkgName
	}
}

func (hw *HuaWeiPush) buildBatchPush(tokens []string) (requestId string, err error) {
	//nspCtx, _ := query.Values()
	//hw.NspCtx.AppId = hw.ClientId
	//fmt.Println(hw.NspCtx)
	nspCtxByt, _ := json.Marshal(hw.NspCtx)
	nspCtxVal := url.Values{}
	nspCtxVal.Add("nsp_ctx", string(nspCtxByt))
	req, err := hw.buildReq(fmt.Sprintf("%s?%s", PRO_API_HW_SEND, nspCtxVal.Encode()))
	if err != nil {
		return
	}
	hw.buildHWBraodCast()
	if len(tokens) > 0 {
		hw.BroadCast.DeviceTokens = make([]string, 0, 0)
		for _, token := range tokens {
			hw.BroadCast.DeviceTokens = append(hw.BroadCast.DeviceTokens, token)
		}
		tokenBytArr, _ := json.Marshal(hw.BroadCast.DeviceTokens)
		hw.BroadCast.DeviceTokenStr = string(tokenBytArr)
	}
	broadCastBytArr, _ := json.Marshal(hw.BroadCast.Payload)
	hw.BroadCast.PayloadStr = string(broadCastBytArr)
	v, _ := query.Values(hw.BroadCast)
	req.Body = []byte(v.Encode())
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := HWPushResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	fmt.Printf("\nresp:%v\n", resp)
	if resp.Code != "80000000" {
		err = errors.New(resp.Msg)
		return
	}
	requestId = resp.RequestId
	return
}

func (hw *HuaWeiPush) push(title, content string, extras map[string]string, tokens []string) (err error) {
	hw.BroadCast.Payload.HPS.Msg.Body.Title = title
	hw.BroadCast.Payload.HPS.Msg.Body.Content = content
	if len(extras) > 0 {
		hw.BroadCast.Payload.HPS.Ext.Customize = make([]map[string]string, 0, 0)
		for k, v := range extras {
			tmp := make(map[string]string, 0)
			tmp[k] = v
			hw.BroadCast.Payload.HPS.Ext.Customize = append(hw.BroadCast.Payload.HPS.Ext.Customize, tmp)
		}
	}
	if len(tokens) > 0 {
		hw.BroadCast.DeviceTokens = make([]string, 0, 0)
		if len(tokens) > 1000 {
			reqLengthMod := len(tokens) % 1000
			reqLoop := len(tokens) / 1000
			if reqLoop > 0 {
				for i := 0; i < reqLoop; i++ {
					start := i * 1000
					tmpTokens := tokens[start : start+1000]
					_, err = hw.buildBatchPush(tmpTokens)
				}
			}
			if reqLengthMod > 0 {
				leftTokens := tokens[reqLoop*1000:]
				_, err = hw.buildBatchPush(leftTokens)
			}
		} else {
			_, err = hw.buildBatchPush(tokens)
		}
	} else {
		_, err = hw.buildBatchPush([]string{})
	}
	return
}

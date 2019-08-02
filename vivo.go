package go_app_push

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"time"
)

const (
	PRO_API_VIVO_PREFIX              string = "https://api-push.vivo.com.cn"
	PRO_API_VIVO_SUBFIX_TOKEN        string = "/message/auth"
	PRO_API_VIVO_SUBFIX_SAVE         string = "/message/saveListPayload"
	PRO_API_VIVO_SUBFIX_BROADCAST    string = "/message/all"
	PRO_API_VIVO_SUBFIX_UNICAST      string = "/message/send"
	PRO_API_VIVO_SUBFIX_UNICASTBATCH string = "/message/pushToList"
)

type (
	VIVOPushType uint32
	VIVONotifyType uint32
	VIVONotifyOpenType uint32
)

const (
	VIVOPushTypeAll            VIVOPushType = 1
	VIVOPushTypeRegistrationId VIVOPushType = 2
	VIVOPushTypeAlias          VIVOPushType = 3
)

const (
	VIVOPushTypeNone            VIVONotifyType = 1
	VIVOPushTypeSound           VIVONotifyType = 2
	VIVOPushTypeVibrate         VIVONotifyType = 3
	VIVOPushTypeSoundAndVibrate VIVONotifyType = 4
)

const (
	VIVOPushTypeAppHome     VIVONotifyOpenType = 1
	VIVOPushTypeUrl         VIVONotifyOpenType = 2
	VIVOPushTypeCustom      VIVONotifyOpenType = 3
	VIVOPushTypeAppActivity VIVONotifyOpenType = 4
)

var pid = uint32(time.Now().UnixNano() % 4294967291)

type VIVOPush struct {
	AuthToken        string            `url:"-" json:"-"`
	TokenCreatedAt   int64             `url:"-" json:"-"`
	AppSecretKey     string            `url:"-" json:"-"`
	IgnoreCheckToken bool              `url:"-" json:"-"`
	Sign             string            `url:"sign,omitempty" json:"sign,omitempty"`
	AppId            int               `url:"appId,omitempty" json:"appId,omitempty"`
	AppKey           string            `url:"appKey,omitempty" json:"appKey,omitempty"`
	Timestamp        int64             `url:"timestamp,omitempty" json:"timestamp,omitempty"`
	Notify           VIVONotifyPayload `url:"-" json:"-"`
}

type VIVOCommonResponse struct {
	Code       int                 `json:"result,omitempty"`
	Message    string              `json:"desc,omitempty"`
	AuthToken  string              `json:"authToken,omitempty"`
	TaskId     string              `json:"taskId,omitempty"`
	Statistics []map[string]string `json:"statistics,omitempty"`
}

type VIVONotifyPayload struct {
	RequestId   string             `url:"requestId,omitempty" json:"requestId,omitempty"`
	RegIds      []string           `url:"regIds,omitempty" json:"regIds,omitempty"`
	Aliases     []string           `url:"aliases,omitempty" json:"aliases,omitempty"`
	TaskId      string             `url:"taskId,omitempty" json:"taskId,omitempty"`
	RegId       string             `url:"regId,omitempty" json:"regId,omitempty"`
	Alias       string             `url:"alias,omitempty" json:"alias,omitempty"`
	Title       string             `url:"title,omitempty" json:"title,omitempty"`
	Content     string             `url:"content,omitempty" json:"content,omitempty"`
	NotifyType  VIVONotifyType     `url:"notifyType,omitempty" json:"notifyType,omitempty"`
	TimeToLive  int                `url:"timeToLive,omitempty" json:"timeToLive,omitempty"`
	SkipType    VIVONotifyOpenType `url:"skipType,omitempty" json:"skipType,omitempty"`
	SkipContent string             `url:"skipContent,omitempty" json:"skipContent,omitempty"`
	NetWorkType int                `url:"networkType,omitempty" json:"networkType,omitempty"` //-1 不限,1:wifi
	Extras      map[string]string  `url:"clientCustomMap,omitempty" json:"clientCustomMap,omitempty"`
	Callback    struct {
		Callback   string `url:"callback,omitempty" json:"callback"`
		Parameters string `url:"callback.param,omitempty" json:"callback.param"`
	} `url:"extra,omitempty" json:"extra,omitempty"`
}

func newVIVOPush() *VIVOPush {
	return new(VIVOPush)
}

func (vo *VIVOPush) checkTokenExpired() {
	if vo.TokenCreatedAt+86400000 < vo.ms() {
		vo.IgnoreCheckToken = true
		err := vo.getToken()
		glog.Info("checkTokenExpired----vo.getTokenErr:%v\n", err)
	}
}

func (vo *VIVOPush) getToken() (err error) {
	vo.sign()
	req, err := vo.buildReq(PRO_API_VIVO_SUBFIX_TOKEN)
	if err != nil {
		return
	}
	v, _ := json.Marshal(vo)
	req.Body = v
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := VIVOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	vo.AuthToken = resp.AuthToken
	vo.TokenCreatedAt = vo.ms()
	vo.IgnoreCheckToken = false
	return
}

func (vo *VIVOPush) SaveNotifyToVIVO() (taskId string, err error) {
	req, err := vo.buildReq(PRO_API_VIVO_SUBFIX_SAVE)
	if err != nil {
		return
	}
	v, _ := json.Marshal(vo.Notify)
	req.Body = v
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := VIVOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	taskId = resp.TaskId
	return
}

func (vo *VIVOPush) PushBroadCast() (taskId string, err error) {
	req, err := vo.buildReq(PRO_API_VIVO_SUBFIX_BROADCAST)
	if err != nil {
		return
	}
	v, _ := json.Marshal(vo.Notify)
	req.Body = v
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := VIVOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	taskId = resp.TaskId
	return
}

func (vo *VIVOPush) PushUniCast() (taskId string, err error) {
	if len(vo.Notify.Alias) == 0 || len(vo.Notify.RegId) == 0 {
		err = VIVOMissingTargetErr
		return
	}
	req, err := vo.buildReq(PRO_API_VIVO_SUBFIX_UNICAST)
	if err != nil {
		return
	}
	v, _ := json.Marshal(vo.Notify)
	req.Body = v
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := VIVOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	taskId = resp.TaskId
	return
}

func (vo *VIVOPush) PushUniBatchCast() (err error) {
	req, err := vo.buildReq(PRO_API_VIVO_SUBFIX_UNICASTBATCH)
	if err != nil {
		return
	}
	vo.Notify.Alias = ""
	vo.Notify.RegId = ""
	if len(vo.Notify.RegIds) == 0 || len(vo.Notify.Aliases) == 0 {
		err = VIVOMissingBatchTargetErr
		return
	}
	var taskId string
	if len(vo.Notify.TaskId) == 0 {
		tmpAliases := vo.Notify.Aliases
		tmpRegids := vo.Notify.RegIds
		vo.Notify.Aliases = make([]string, 0, 0)
		vo.Notify.RegIds = make([]string, 0, 0)
		taskId, err = vo.SaveNotifyToVIVO()
		if err != nil {
			return
		}
		vo.Notify.TaskId = taskId
		vo.Notify.RegIds = tmpRegids
		vo.Notify.Aliases = tmpAliases
	}
	v, _ := json.Marshal(vo.Notify)
	req.Body = v
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := VIVOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	return
}

func (vo *VIVOPush) push(title, content string, extras map[string]string, tokens []string) (err error) {
	vo.Notify.Title = title
	vo.Notify.Content = content
	if len(extras) > 0 {
		vo.Notify.Extras = extras
	}
	if len(tokens) == 0 {
		_, err = vo.PushBroadCast()
	} else if len(tokens) > 1 {
		if len(tokens) > 1000 {
			reqLengthMod := len(tokens) % 1000
			reqLoop := len(tokens) / 1000
			if reqLoop > 0 {
				for i := 0; i < reqLoop; i++ {
					start := i * 1000
					tmpTokens := tokens[start : start+1000]
					err = vo.buildBatchPush(tmpTokens)
				}
			}
			if reqLengthMod > 0 {
				leftTokens := tokens[reqLoop*1000:]
				err = vo.buildBatchPush(leftTokens)
			}
		} else {
			err = vo.buildBatchPush(tokens)
		}
	} else if len(tokens) == 1 {
		vo.Notify.Alias = tokens[0]
		_, err = vo.PushUniCast()
	}
	return
}

func (vo *VIVOPush) buildReq(queryPath string) (req *PushReq, err error) {
	if vo.AppId == 0 {
		err = VIVOMissingAppIdErr
		return
	}
	if len(vo.AppKey) == 0 {
		err = VIVOMissingAppKeyErr
		return
	}
	if len(vo.AppSecretKey) == 0 {
		err = VIVOMissingAppSecretKeyErr
		return
	}
	if !vo.IgnoreCheckToken {
		vo.checkTokenExpired()
	}
	if len(vo.Notify.RequestId) == 0 {
		vo.requestId()
	}
	req = newPushReq()
	req.Headers = make(map[string]string, 0)
	req.Headers["authToken"] = vo.AuthToken
	req.Headers["Content-Type"] = "application/json"
	req.Method = "POST"
	req.Url = fmt.Sprintf("%s%s", PRO_API_VIVO_PREFIX, queryPath)
	return
}

func (vo *VIVOPush) buildBatchPush(tokens []string) (err error) {
	vo.Notify.Aliases = tokens
	return vo.PushUniBatchCast()
}

func (vo *VIVOPush) sign() {
	vo.Timestamp = vo.ms()
	vo.Sign = vo.md5(fmt.Sprintf("%d%s%d%s", vo.AppId, vo.AppKey, vo.Timestamp, vo.AppSecretKey))
}

func (vo *VIVOPush) requestId() {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	vo.Notify.RequestId = vo.md5(base64.URLEncoding.EncodeToString(b[:]))
}

func (vo *VIVOPush) ms() int64 {
	return time.Now().UnixNano() / 1000000
}

func (vo *VIVOPush) md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

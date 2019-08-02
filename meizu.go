// api document
// http://open-wiki.flyme.cn/doc-wiki/index
package go_app_push

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"strings"
)

const (
	PRO_API_MZ_PREFIX                string = "http://server-api-mzups.meizu.com"
	PRO_API_MZ_MSG_TRANSPARENT       string = "/ups/api/server/push/unvarnished/pushByPushId"
	PRO_API_MZ_MSG_NOTIFY            string = "/ups/api/server/push/varnished/pushByPushId"
	PRO_API_MZ_MSG_TRANSPARENT_ALIAS string = "/ups/api/server/push/unvarnished/pushByAlias"
	PRO_API_MZ_MSG_NOTIFY_ALIAS      string = "/ups/api/server/push/varnished/pushByAlias"
	PRO_API_MZ_MSG_NOTIFY_ALL        string = "/ups/api/server/push/pushTask/pushToApp"
	PRO_API_MZ_MSG_CANCEL            string = "/ups/api/server/push/pushTask/cancel"
	PRO_API_MZ_MSG_STATICS           string = "/ups/api/server/push/statistics/dailyPushStatics"
)

type MeiZuResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"message"`
	Redirect string `json:"redirect,omitempty"`
	Data     struct {
		AppId      int                 `json:"appId,omitempty"`
		TaskId     int                 `json:"taskId,omitempty"`
		PushType   int                 `json:"pushType,omitempty"`
		MsgId      string              `json:"msgId,omitempty"`
		RespTarget map[string][]string `json:"respTarget,omitempty"`
		Logs       map[string]string   `json:"logs,omitempty"`
	} `json:"value"`
}

type MeiZuPush struct {
	AppId  int
	AppKey string
	Notify MeiZuNotify
}

type PushTimeInfo struct {
	Offline   int `json:"offline,omitempty"`
	ValidTime int `json:"validTime"`
}

type MeiZuMsgTransparent struct {
	Content      string       `json:"content,omitempty"`
	PushTimeInfo PushTimeInfo `json:"pushTimeInfo,omitempty"`
}

type MeiZuMsgNotification struct {
	PushTimeInfo  PushTimeInfo `json:"pushTimeInfo,omitempty"`
	NoticeBarInfo struct {
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"noticeBarInfo,omitempty"`
	ClickTypeInfo struct {
		ClickType  int    `json:"clickType"`
		Url        string `json:"url,omitempty"`
		Parameters string `json:"parameters,omitempty"`
		Activity   string `json:"activity,omitempty"`
	} `json:"clickTypeInfo,omitempty"`
	AdvanceInfo struct {
		Suspend             int    `json:"suspend,omitempty"`
		ClearNoticeBar      int    `json:"clearNoticeBar,omitempty"`
		FixDisplay          int    `json:"fixDisplay,omitempty"`
		FixStartDisplayTime string `json:"fixStartDisplayTime,omitempty"` //0000-00-00 00:00:00
		FixEndDisplayTime   string `json:"fixStartDisplayTime,omitempty"` //0000-00-00 00:00:00
		NotificationType    struct {
			Vibrate int `json:"vibrate,omitempty"`
			Lights  int `json:"lights,omitempty"`
			Sound   int `json:"sound,omitempty"`
		} `json:"notificationType,omitempty"`
	} `json:"advanceInfo,omitempty"`
}

type MeiZuNotify struct {
	AppId           string               `url:"appId,omitempty" json:"appId,omitempty"`
	Sign            string               `url:"sign,omitempty" json:"sign,omitempty"`
	Alias           string               `url:"alias,omitempty" json:"alias,omitempty"`     //逗号分隔,1000单次限制
	PushIds         string               `url:"pushIds,omitempty" json:"pushIds,omitempty"` //逗号分隔,1000单次限制
	PushType        int                  `url:"pushType" json:"pushType"`
	MessageJson     string               `url:"messageJson,omitempty" json:"messageJson,omitempty"`
	MsgTransparent  MeiZuMsgTransparent  `url:"-" json:"-"`
	MsgNotification MeiZuMsgNotification `url:"-" json:"-"`
}

func newMeiZuPush() *MeiZuPush {
	mz := new(MeiZuPush)
	return mz
}

func (mz *MeiZuPush) buildReq(url string) (req *PushReq, err error) {
	if mz.AppId == 0 {
		err = MissingMeiZuAppKeyErr
		return
	}
	if len(mz.AppKey) == 0 {
		err = MissingAppKeyErr
		return
	}
	req = newPushReq()
	req.Headers = make(map[string]string, 0)
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"
	req.Method = "POST"
	req.Url = fmt.Sprintf("%s%s", PRO_API_MZ_PREFIX, url)
	return
}

func (mz *MeiZuPush) PushBroadCast() (resp MeiZuResponse, err error) {
	req, err := mz.buildReq(PRO_API_MZ_MSG_NOTIFY_ALL)
	if err != nil {
		return
	}
	v, _ := json.Marshal(mz.Notify.MsgNotification)
	mz.Notify.MessageJson = string(v)
	queryVal, _ := query.Values(mz.Notify)
	mz.sign(queryVal.Encode())
	queryVal, _ = query.Values(mz.Notify)
	postBodyStr := queryVal.Encode()
	req.Body = []byte(postBodyStr)
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code != 200 {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func (mz *MeiZuPush) PushUniBatchCast() (resp MeiZuResponse, err error) {
	req, err := mz.buildReq(PRO_API_MZ_MSG_NOTIFY_ALIAS)
	if err != nil {
		return
	}
	v, _ := json.Marshal(mz.Notify.MsgNotification)
	mz.Notify.MessageJson = string(v)
	queryVal, _ := query.Values(mz.Notify)
	queryStr := queryVal.Encode()
	queryStr = strings.Replace(queryStr, "pushType=0", "", -1) //unset pushType
	mz.sign(queryStr)
	queryVal, _ = query.Values(mz.Notify)
	postBodyStr := queryVal.Encode()
	req.Body = []byte(postBodyStr)
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code != 200 {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func (mz *MeiZuPush) push(title, content string, extras map[string]string, tokens []string) (err error) {
	mz.Notify.MsgNotification.NoticeBarInfo.Title = title
	mz.Notify.MsgNotification.NoticeBarInfo.Content = content
	mz.Notify.MsgNotification.ClickTypeInfo.ClickType = 0
	if len(extras) > 0 {
		extraBytArr, _ := json.Marshal(extras)
		mz.Notify.MsgNotification.ClickTypeInfo.Parameters = string(extraBytArr)
	}
	if len(tokens) == 0 {
		_, err = mz.PushBroadCast()
	} else {
		if len(tokens) > 1000 {
			reqLengthMod := len(tokens) % 1000
			reqLoop := len(tokens) / 1000
			if reqLoop > 0 {
				for i := 0; i < reqLoop; i++ {
					start := i * 1000
					tmpTokens := tokens[start : start+1000]
					_, err = mz.buildBatchPush(tmpTokens)
				}
			}
			if reqLengthMod > 0 {
				leftTokens := tokens[reqLoop*1000:]
				_, err = mz.buildBatchPush(leftTokens)
			}
		} else {
			_, err = mz.buildBatchPush(tokens)
		}
	}
	return
}

func (mz *MeiZuPush) buildBatchPush(tokens []string) (resp MeiZuResponse, err error) {
	mz.Notify.Alias = strings.Join(tokens, ",")
	return mz.PushUniBatchCast()
}

func (mz *MeiZuPush) sign(param string) {
	param = strings.Replace(param, "&", "", -1)
	mz.Notify.Sign = mz.md5(fmt.Sprintf("%s%s", param, mz.AppKey))
}

func (mz *MeiZuPush) md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

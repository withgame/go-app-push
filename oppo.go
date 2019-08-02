package go_app_push

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/go-querystring/query"
	"strings"
	"time"
)

const (
	PRO_API_OPPO_PREFIX              string = "https://api.push.oppomobile.com/server/v1"
	PRO_API_OPPO_SUBFIX_TOKEN        string = "/auth"
	PRO_API_OPPO_SUBFIX_SAVE         string = "/message/notification/save_message_content"
	PRO_API_OPPO_SUBFIX_BROADCAST    string = "/message/notification/broadcast"
	PRO_API_OPPO_SUBFIX_UNICAST      string = "/message/notification/unicast"
	PRO_API_OPPO_SUBFIX_UNICASTBATCH string = "/message/notification/unicast_batch"
)

type OPPOPushType uint32

const (
	OPPOPushTypeNil            OPPOPushType = 0
	OPPOPushTypeAll            OPPOPushType = 1
	OPPOPushTypeRegistrationId OPPOPushType = 2
	OPPOPushTypeAlias          OPPOPushType = 3
)

type OPPOClickType int

const (
	OPPOClickTypeStartup        OPPOClickType = 0
	OPPOClickTypeActivityAction OPPOClickType = 1
	OPPOClickTypeWeb            OPPOClickType = 2
	OPPOClickTypeActivity       OPPOClickType = 4
	OPPOClickTypeSchemeUrl      OPPOClickType = 5
)

type OPPOPush struct {
	AuthToken        string              `url:"-" json:"-"`
	TokenCreatedAt   int64               `url:"-" json:"-"`
	MasterKey        string              `url:"-" json:"-"`
	IgnoreCheckToken bool                `url:"-" json:"-"`
	Sign             string              `url:"sign,omitempty" json:"sign"`
	AppKey           string              `url:"app_key,omitempty" json:"app_key"`
	Timestamp        int64               `url:"timestamp,omitempty" json:"timestamp"`
	Notify           OPPONotifyPayload   `url:"-" json:"-"`
	Broadcast        BroadCastPayload    `url:"-" json:"-"`
	Unicast          UniCastPayload      `url:"-" json:"-"`
	UniBatchcast     UniBatchCastPayload `url:"-" json:"-"`
	PushType         OPPOPushType        `url:"-" json:"-"`
}

type OPPOCommonResponse struct {
	Code    int                    `json:"code,omitempty"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type OPPOBroadCastResponse struct {
	Code    int                    `json:"code,omitempty"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data"`
}

type UniCastPayload struct {
	Message string             `url:"message" json:"message"`
	Payload UniCastPayloadBody `url:"-" json:"-"`
}

type UniBatchCastPayload struct {
	Message string               `url:"messages" json:"messages"`
	Payload []UniCastPayloadBody `url:"-" json:"-"`
}

type BroadCastPayload struct {
	MessageId   string       `url:"message_id,omitempty" json:"message_id,omitempty"`
	TargetType  OPPOPushType `url:"target_type,omitempty" json:"target_type,omitempty"`
	TargetValue string       `url:"target_value,omitempty" json:"target_value,omitempty"`
}

type UniCastPayloadBody struct {
	TargetType     OPPOPushType      `url:"target_type,omitempty" json:"target_type,omitempty"`
	TargetValue    string            `url:"target_value,omitempty" json:"target_value,omitempty"`
	RegistrationId string            `url:"registration_id,omitempty" json:"registration_id,omitempty"`
	Notification   string            `url:"notification,omitempty" json:"notification,omitempty"`
	Notify         OPPONotifyPayload `url:"-" json:"-"`
}

type OPPONotifyPayload struct {
	AppMessageId        string        `url:"app_message_id" json:"app_message_id,omitempty"`
	Title               string        `url:"title" json:"title"`
	SubTitle            string        `url:"sub_title" json:"sub_title"`
	Content             string        `url:"content" json:"content"`
	ClickActionType     OPPOClickType `url:"click_action_type,omitempty" json:"click_action_type,omitempty"`
	ClickActionActivity string        `url:"click_action_activity,omitempty" json:"click_action_activity,omitempty"`
	ClickActionUrl      string        `url:"click_action_url,omitempty" json:"click_action_url,omitempty"`
	ActionParameters    string        `url:"action_parameters,omitempty" json:"action_parameters,omitempty"`
	ShowTimeType        int           `url:"show_time_type,omitempty" json:"show_time_type,omitempty"`
	ShowStartTime       int64         `url:"show_start_time,omitempty" json:"show_start_time,omitempty"`
	ShowEndTime         int64         `url:"show_end_time,omitempty" json:"show_end_time,omitempty"`
	OffLine             bool          `url:"off_line,omitempty" json:"off_line,omitempty"`
	OffLineTtl          int           `url:"off_line_ttl,omitempty" json:"off_line_ttl,omitempty"`
	PushTimeType        int           `url:"push_time_type,omitempty" json:"push_time_type,omitempty"`
	PushStartTime       int64         `url:"push_start_time,omitempty" json:"push_start_time,omitempty"`
	TimeZone            string        `url:"time_zone,omitempty" json:"time_zone,omitempty"`
	FixSpeed            bool          `url:"fix_speed,omitempty" json:"fix_speed,omitempty"`
	FixSpeedRate        int64         `url:"fix_speed_rate,omitempty" json:"fix_speed_rate,omitempty"`
	NetWorkType         int           `url:"net_work_type,omitempty" json:"net_work_type,omitempty"`
	CallBackUrl         string        `url:"call_back_url,omitempty" json:"call_back_url,omitempty"`
	CallBackParameter   string        `url:"call_back_parameter,omitempty" json:"call_back_parameter,omitempty"`
}

func newOPPOPush() *OPPOPush {
	return new(OPPOPush)
}

/**
 * check oppo api token
 */
func (op *OPPOPush) checkTokenExpired() {
	if op.TokenCreatedAt+86400000 < op.ms() {
		op.IgnoreCheckToken = true
		err := op.getToken()
		glog.Info("checkTokenExpired----op.getTokenErr:%v\n", err)
	}
}

/**
 * get oppo api token
 */
func (op *OPPOPush) getToken() (err error) {
	op.sign()
	req, err := op.buildReq(PRO_API_OPPO_SUBFIX_TOKEN)
	if err != nil {
		return
	}
	v, _ := query.Values(op)
	req.Body = []byte(v.Encode())
	body, statusCode, _, err := req.doPushRequest()
	// fmt.Printf("body:%v\n", string(body))
	// fmt.Printf("statusCode:%v\n", statusCode)
	// fmt.Printf("getTokenStatusCodeErr:%v\n", err)
	if statusCode != 200 {
		return
	}
	// fmt.Printf("req.doPushRequestErr:%v\n", err)
	if err != nil {
		return
	}
	resp := OPPOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	// fmt.Printf("OPPOCommonResponseResp:%v\n", resp)
	// fmt.Printf("OPPOCommonResponseRespErr:%v\n", err)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	// fmt.Printf("getTokenResp:%v\n", resp)
	//op.AuthToken, ok1 := resp.Data["auth_token"].(string)
	if atoken, ok1 := resp.Data["auth_token"].(string); ok1 {
		op.AuthToken = atoken
	}
	if ctime, ok2 := resp.Data["create_time"].(int64); ok2 {
		op.TokenCreatedAt = ctime
	}
	op.IgnoreCheckToken = false
	// fmt.Printf("OPPOPushGetTokenOPPOresp.Data------->:%v\n", resp)
	// fmt.Printf("OPPOPushGetTokenOPPO------->:%v\n", op)
	return
}

func (op *OPPOPush) SaveNotifyToOPPO() (msgId string, err error) {
	req, err := op.buildReq(PRO_API_OPPO_SUBFIX_SAVE)
	if err != nil {
		return
	}
	v, _ := query.Values(op.Notify)
	req.Body = []byte(v.Encode())
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := OPPOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	if msgIdVal, ok := resp.Data["message_id"].(string); ok {
		msgId = msgIdVal
	}
	return
}

func (op *OPPOPush) PushBroadCast() (msgId, taskId string, err error) {
	msgId, _ = op.SaveNotifyToOPPO()
	req, err := op.buildReq(PRO_API_OPPO_SUBFIX_BROADCAST)
	if err != nil {
		return
	}
	op.Broadcast.MessageId = msgId
	op.Broadcast.TargetType = OPPOPushTypeAll
	v, _ := query.Values(op.Broadcast)
	req.Body = []byte(v.Encode())
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := OPPOBroadCastResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	if len(resp.Data) > 2 {
		for k, v := range resp.Data {
			if k == "message_id" {
				msgId = v.(string)
			} else if k == "task_id" {
				taskId = v.(string)
			}
		}
	} else {
		msgId = resp.Data["message_id"].(string)
		taskId = resp.Data["task_id"].(string)
	}
	return
}

func (op *OPPOPush) PushUniCast() (msgId string, err error) {
	req, err := op.buildReq(PRO_API_OPPO_SUBFIX_UNICAST)
	if err != nil {
		return
	}
	if op.PushType == OPPOPushTypeNil {
		err = OPPOMissingPushTypeErr
		return
	}
	var tmpActionParams string
	if len(op.Notify.ActionParameters) > 0 {
		tmpActionParams = op.Notify.ActionParameters
		op.Notify.ActionParameters = "--aparams--"
		// fmt.Printf("tmpActionParams:%s\n\n", tmpActionParams)
	}
	op.Unicast.Payload.Notify = op.Notify
	if len(op.Unicast.Payload.TargetValue) == 0 {
		err = OPPOMissingDeviceErr
		return
	}
	if op.Unicast.Payload.TargetType == 0 {
		op.Unicast.Payload.TargetType = 3
	}
	notifyByteArr, _ := json.Marshal(op.Unicast.Payload.Notify)
	notifyStr := string(notifyByteArr)
	// fmt.Printf("notifyStr:%s\n", notifyStr)

	op.Unicast.Payload.Notification = "--notify--"
	notifyPayloadByteArr, _ := json.Marshal(op.Unicast.Payload)

	notifyPayloadStr := string(notifyPayloadByteArr)

	// fmt.Printf("notifyPayloadByteArr:%s\n", notifyPayloadStr)

	op.Unicast.Message = strings.Replace(notifyPayloadStr, "\"--notify--\"", notifyStr, -1)

	if len(op.Notify.ActionParameters) > 0 {
		op.Unicast.Message = strings.Replace(op.Unicast.Message, "\"--aparams--\"", tmpActionParams, -1)
	}

	// fmt.Printf("op.Unicast.Message---->:%s\n\n", op.Unicast.Message)

	v, _ := query.Values(op.Unicast)

	req.Body = []byte(v.Encode())

	// fmt.Printf("OPPOPushPushUniCast--req--->:%v\n", req)
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := OPPOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	//msgId = resp.Data["messageId"]
	if msgIdVal, ok := resp.Data["messageId"].(string); ok {
		msgId = msgIdVal
	}
	// fmt.Printf("PushUniCast---msgId----->:%s\n", msgId)
	return
}

func (op *OPPOPush) PushUniBatchCast() (msgId string, err error) {
	req, err := op.buildReq(PRO_API_OPPO_SUBFIX_UNICASTBATCH)
	if err != nil {
		return
	}
	if len(op.UniBatchcast.Payload) == 0 {
		return
	}
	var notifyActionParamsStr string
	if len(op.Notify.ActionParameters) > 0 {
		notifyActionParamsStr = op.Notify.ActionParameters
		op.Notify.ActionParameters = "--aprams--"
	}

	tmpPayloadNotifiesStr := make(map[string]string, 0)
	for index, payload := range op.UniBatchcast.Payload {
		payload.Notify = op.Notify
		tmpNotifyByteArr, _ := json.Marshal(payload.Notify)
		tmpNotifyStr := string(tmpNotifyByteArr)
		keyStr := fmt.Sprintf("--%d-notify--", index)
		tmpPayloadNotifiesStr[keyStr] = tmpNotifyStr
		op.UniBatchcast.Payload[index].Notification = keyStr
	}
	uniBatchCastByteArr, _ := json.Marshal(op.UniBatchcast.Payload)
	batchCastStr := string(uniBatchCastByteArr)

	for index, _ := range op.UniBatchcast.Payload {
		keyStr := fmt.Sprintf("--%d-notify--", index)
		keyReplaceStr := fmt.Sprintf("\"--%d-notify--\"", index)
		batchCastStr = strings.Replace(batchCastStr, keyReplaceStr, tmpPayloadNotifiesStr[keyStr], -1)
	}
	if len(notifyActionParamsStr) > 0 {
		batchCastStr = strings.Replace(batchCastStr, "\"--aprams--\"", notifyActionParamsStr, -1)
	}
	op.UniBatchcast.Message = batchCastStr
	v, _ := query.Values(op.UniBatchcast)
	req.Body = []byte(v.Encode())
	body, _, _, err := req.doPushRequest()
	if err != nil {
		return
	}
	resp := OPPOCommonResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	if resp.Code > 0 {
		err = errors.New(resp.Message)
		return
	}
	//msgId = resp.Data["messageId"]
	if msgIdVal, ok := resp.Data["messageId"].(string); ok {
		msgId = msgIdVal
	}
	return
}

/**
 * start push action
 */
func (op *OPPOPush) push(title, content string, extras map[string]string, tokens []string) (err error) {
	op.Notify.Title = title
	op.Notify.Content = content
	if len(extras) > 0 {
		extrasBytArr, _ := json.Marshal(extras)
		op.Notify.ActionParameters = string(extrasBytArr)
	}
	if len(tokens) == 0 {
		_, _, err = op.PushBroadCast()
	} else if len(tokens) > 1 {
		if len(tokens) > 1000 {
			reqLengthMod := len(tokens) % 1000
			reqLoop := len(tokens) / 1000
			if reqLoop > 0 {
				for i := 0; i < reqLoop; i++ {
					start := i * 1000
					tmpTokens := tokens[start : start+1000]
					_, err = op.buildBatchPush(tmpTokens)
				}
			}
			if reqLengthMod > 0 {
				leftTokens := tokens[reqLoop*1000:]
				_, err = op.buildBatchPush(leftTokens)
			}
		} else {
			_, err = op.buildBatchPush(tokens)
		}
	} else if len(tokens) == 1 {
		//op.Unicast.Payload.TargetType = OPPOPushTypeAlias
		op.Unicast.Payload.TargetType = op.PushType
		op.Unicast.Payload.TargetValue = tokens[0]
		_, err = op.PushUniCast()
	}
	return
}

func (op *OPPOPush) buildReq(queryPath string) (req *PushReq, err error) {
	if len(op.AppKey) == 0 {
		err = MissingAppKeyErr
		return
	}
	if len(op.MasterKey) == 0 {
		err = OPPOMissingMasterKeyErr
		return
	}
	if !op.IgnoreCheckToken {
		op.checkTokenExpired()
	}
	req = newPushReq()
	req.Headers = make(map[string]string, 0)
	req.Headers["auth_token"] = op.AuthToken
	req.Method = "POST"
	req.Url = fmt.Sprintf("%s%s", PRO_API_OPPO_PREFIX, queryPath)
	// glog.Infof("OPPOPush) buildReq---->req:%v\n", req)
	return
}

func (op *OPPOPush) buildBatchPush(tokens []string) (msgId string, err error) {
	if op.PushType == OPPOPushTypeNil {
		err = OPPOMissingPushTypeErr
		return
	}
	op.UniBatchcast.Payload = make([]UniCastPayloadBody, 0, 0)
	for _, token := range tokens {
		p := UniCastPayloadBody{}
		p.Notify = op.Notify
		//p.TargetType = OPPOPushTypeAlias
		p.TargetType = op.PushType
		p.TargetValue = token
		op.UniBatchcast.Payload = append(op.UniBatchcast.Payload, p)
	}
	return op.PushUniBatchCast()
}

func (op *OPPOPush) sign() {
	op.Timestamp = op.ms()
	signStr := fmt.Sprintf("%s%d%s", op.AppKey, op.Timestamp, op.MasterKey)
	// fmt.Printf("OPPOPushBeforeSignStr:%s", signStr)
	op.Sign = op.sha256(signStr)
	// fmt.Printf("OPPOPushSignStr:%s\n", op.Sign)
}

func (op *OPPOPush) ms() int64 {
	return time.Now().UnixNano() / 1000000
}

func (op *OPPOPush) sha256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

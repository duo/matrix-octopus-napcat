package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type PacketType string

const (
	PktRequest  PacketType = "request"
	PktResponse PacketType = "response"
	PktNotice   PacketType = "notice"
	PktMessage  PacketType = "message"
)

type Packet struct {
	ID      int64      `json:"id"`
	MXID    string     `json:"mxid"`
	Type    PacketType `json:"type"`
	Payload any        `json:"payload,omitempty"`
}

type MethodType string

const (
	MethodConnect            MethodType = "connect"
	MethodDisconnect         MethodType = "disconnect"
	MethodLogin              MethodType = "login"
	MethodLogout             MethodType = "logout"
	MethodGetLoginInfo       MethodType = "get_login_info"
	MethodGetUserInfo        MethodType = "get_user_info"
	MethodGetGroupInfo       MethodType = "get_group_info"
	MethodGetFriendList      MethodType = "get_friend_list"
	MethodGetGroupList       MethodType = "get_group_list"
	MethodGetGroupMemberList MethodType = "get_group_member_list"
	MethodGetGroupMemberInfo MethodType = "get_group_member_info"
	MethodSendMessage        MethodType = "send_message"
	MethodRevokeMessage      MethodType = "revoke_message"
)

type LoginStepType string

const (
	LoginStepTypeUserInput      LoginStepType = "user_input"
	LoginStepTypeDisplayAndWait LoginStepType = "display_and_wait"
	LoginStepTypeWait           LoginStepType = "wait"
	LoginStepTypeComplete       LoginStepType = "complete"
)

type LoginDisplayType string

const (
	LoginDisplayTypeQR   LoginDisplayType = "qr"
	LoginDisplayTypeCode LoginDisplayType = "code"
)

type LoginDisplayAndWaitParams struct {
	Type     LoginDisplayType `json:"type"`
	Data     string           `json:"data"`
	ImageURL string           `json:"image_url,omitempty"`
}

type LoginInputFieldType string

const (
	LoginInputFieldTypeUsername    LoginInputFieldType = "username"
	LoginInputFieldTypePassword    LoginInputFieldType = "password"
	LoginInputFieldTypePhoneNumber LoginInputFieldType = "phone_number"
	LoginInputFieldTypeEmail       LoginInputFieldType = "email"
	LoginInputFieldType2FACode     LoginInputFieldType = "2fa_code"
	LoginInputFieldTypeToken       LoginInputFieldType = "token"
	LoginInputFieldTypeURL         LoginInputFieldType = "url"
	LoginInputFieldTypeDomain      LoginInputFieldType = "domain"
)

type LoginInputDataField struct {
	Type        LoginInputFieldType `json:"type"`
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Pattern     string              `json:"pattern,omitempty"`
}

type LoginUserInputParams struct {
	Fields []LoginInputDataField `json:"fields"`
}

type LoginCompleteParams struct {
	LoginInfo *UserInfo `json:"login_info,omitempty"`
}

type LoginStep struct {
	StepID       string        `json:"id"`
	Type         LoginStepType `json:"type"`
	Instructions string        `json:"instructions"`

	DisplayAndWaitParams *LoginDisplayAndWaitParams `json:"display_and_wait,omitempty"`
	UserInputParams      *LoginUserInputParams      `json:"user_input,omitempty"`
	CompleteParams       *LoginCompleteParams       `json:"complete,omitempty"`
}

type GetLoginQRData struct {
	Image []byte `json:"image"`
}

type GetFriendListData struct {
	Friends []*UserInfo `json:"friends"`
}

type GetGroupListData struct {
	Groups []*GroupInfo
}

type ChatType int

const (
	ChatUnknown ChatType = 0
	ChatPrivate ChatType = 1
	ChatGroup   ChatType = 2
)

var (
	chatTypeMap = map[string]ChatType{
		"unknown": ChatUnknown,
		"private": ChatPrivate,
		"group":   ChatGroup,
	}
)

func (c ChatType) IsPrivate() bool {
	return c == ChatPrivate
}

func ParseChatTypeStr(str string) ChatType {
	if c, ok := chatTypeMap[strings.ToLower(str)]; ok {
		return c
	}
	return ChatUnknown
}

type Request struct {
	Type   MethodType `json:"type"`
	Params any        `json:"params,omitempty"`
}

type Response struct {
	Type  MethodType     `json:"type"`
	Error *ErrorResponse `json:"error,omitempty"`
	Data  any            `json:"data,omitempty"`
}

type ErrorResponse struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

type NoticeType string

const (
	NoticeHeartbeat NoticeType = "heartbeat"
)

type Heartbeat struct {
	IsLoggedIn bool   `json:"is_logged_in"`
	Interval   uint32 `josn:"interval"`
}

type Notice struct {
	ClientID  string     `json:"client_id"`
	Timestamp int64      `json:"timestamp"`
	Type      NoticeType `json:"type"`
	Data      any        `json:"data,omitempty"`
}

type UserInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Alias  string `json:"alias,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

type GroupInfo struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Alias   string   `json:"alias,omitempty"`
	Avatar  string   `json:"avatar,omitempty"`
	Notice  string   `json:"notice,omitempty"`
	Members []string `json:"members"`
}

type MemberInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Alias  string `json:"alias,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

type MessageType string

const (
	MsgText     MessageType = "text"
	MsgImage    MessageType = "image"
	MsgSticker  MessageType = "sticker"
	MsgAudio    MessageType = "audio"
	MsgVideo    MessageType = "video"
	MsgFile     MessageType = "file"
	MsgLocation MessageType = "location"
	MsgNotice   MessageType = "notice"
	MsgApp      MessageType = "app"
	MsgRevoke   MessageType = "revoke"
	MsgSystem   MessageType = "system"
)

type Message struct {
	ID         string      `json:"id"`
	ThreadID   string      `json:"thread_id,omitempty"`
	Timestamp  int64       `json:"timestamp"`
	From       User        `json:"from"`
	Chat       Chat        `json:"chat"`
	Type       MessageType `json:"type"`
	Content    string      `json:"content,omitempty"`
	Mentions   []string    `json:"mentions,omitempty"`
	Reply      *ReplyInfo  `json:"reply,omitempty"`
	Attachment any         `json:"attachment,omitempty"`
}

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Alias  string `json:"alias,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

type Chat struct {
	ID    string   `json:"id"`
	Type  ChatType `json:"type"`
	Title string   `json:"title,omitempty"`
}

type ReplyInfo struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Content   string `json:"content,omitempty"`
}

type AppData struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"desc,omitempty"`
	URL         string `json:"url,omitempty"`
}

type LocationData struct {
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type BlobData struct {
	Name   string `json:"name,omitempty"`
	Mime   string `json:"mime,omitempty"`
	Binary []byte `json:"binary"`
}

type MessageRevoke struct {
	Chat      Chat   `json:"chat"`
	Opeartor  User   `json:"operator"`
	MessageID string `json:"message_id"`
}

func (o *Packet) UnmarshalJSON(data []byte) error {
	type cloneType Packet

	rawMsg := json.RawMessage{}
	o.Payload = &rawMsg

	if err := json.Unmarshal(data, (*cloneType)(o)); err != nil {
		return err
	}

	switch o.Type {
	case PktRequest:
		var request *Request
		if err := json.Unmarshal(rawMsg, &request); err != nil {
			return err
		}
		o.Payload = request
	case PktResponse:
		var response *Response
		if err := json.Unmarshal(rawMsg, &response); err != nil {
			return err
		}
		o.Payload = response
	case PktNotice:
		var notice *Notice
		if err := json.Unmarshal(rawMsg, &notice); err != nil {
			return err
		}
		o.Payload = notice
	case PktMessage:
		var msg *Message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			return err
		}
		o.Payload = msg
	}

	return nil
}

func (o *Request) UnmarshalJSON(data []byte) error {
	type cloneType Request

	rawMsg := json.RawMessage{}
	o.Params = &rawMsg

	if err := json.Unmarshal(data, (*cloneType)(o)); err != nil {
		return err
	}

	switch o.Type {
	case MethodLogin:
		var params *LoginStep
		if err := json.Unmarshal(rawMsg, &params); err != nil {
			return err
		}
		o.Params = params
	case MethodGetUserInfo, MethodGetGroupInfo, MethodGetGroupMemberList, MethodGetGroupMemberInfo, MethodRevokeMessage:
		var params []string
		if err := json.Unmarshal(rawMsg, &params); err != nil {
			return err
		}
		o.Params = params
	case MethodSendMessage:
		var params *Message
		if err := json.Unmarshal(rawMsg, &params); err != nil {
			return err
		}
		o.Params = params
	}

	return nil
}

func (o *Response) UnmarshalJSON(data []byte) error {
	type cloneType Response

	rawMsg := json.RawMessage{}
	o.Data = &rawMsg

	if err := json.Unmarshal(data, (*cloneType)(o)); err != nil {
		return err
	}

	if o.Error != nil {
		return nil
	}

	switch o.Type {
	case MethodLogin:
		var step *LoginStep
		if err := json.Unmarshal(rawMsg, &step); err != nil {
			return err
		}
		o.Data = step
	case MethodGetLoginInfo, MethodGetUserInfo:
		var info *UserInfo
		if err := json.Unmarshal(rawMsg, &info); err != nil {
			return err
		}
		o.Data = info
	case MethodGetGroupInfo:
		var info *GroupInfo
		if err := json.Unmarshal(rawMsg, &info); err != nil {
			return err
		}
		o.Data = info
	case MethodGetFriendList:
		var friends []*UserInfo
		if err := json.Unmarshal(rawMsg, &friends); err != nil {
			return err
		}
		o.Data = friends
	case MethodGetGroupList:
		var groups []*GroupInfo
		if err := json.Unmarshal(rawMsg, &groups); err != nil {
			return err
		}
		o.Data = groups
	case MethodGetGroupMemberList:
		var members []*MemberInfo
		if err := json.Unmarshal(rawMsg, &members); err != nil {
			return err
		}
		o.Data = members
	case MethodGetGroupMemberInfo:
		var info *MemberInfo
		if err := json.Unmarshal(rawMsg, &info); err != nil {
			return err
		}
		o.Data = info
	case MethodSendMessage:
		var msg *Message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			return err
		}
		o.Data = msg
	}

	return nil
}

func (o *Notice) UnmarshalJSON(data []byte) error {
	type cloneType Notice

	rawMsg := json.RawMessage{}
	o.Data = &rawMsg

	if err := json.Unmarshal(data, (*cloneType)(o)); err != nil {
		return err
	}

	switch o.Type {
	case NoticeHeartbeat:
		var heartbeat *Heartbeat
		if err := json.Unmarshal(rawMsg, &heartbeat); err != nil {
			return err
		}
		o.Data = heartbeat
	}

	return nil
}

func (o *Message) UnmarshalJSON(data []byte) error {
	type cloneType Message

	rawMsg := json.RawMessage{}
	o.Attachment = &rawMsg

	if err := json.Unmarshal(data, (*cloneType)(o)); err != nil {
		return err
	}

	switch o.Type {
	case MsgImage, MsgSticker, MsgAudio, MsgVideo, MsgFile:
		var blobs []*BlobData
		if err := json.Unmarshal(rawMsg, &blobs); err != nil {
			return err
		}
		o.Attachment = blobs
	case MsgLocation:
		var location *LocationData
		if err := json.Unmarshal(rawMsg, &location); err != nil {
			return err
		}
		o.Attachment = location
	case MsgApp:
		var app *AppData
		if err := json.Unmarshal(rawMsg, &app); err != nil {
			return err
		}
		o.Attachment = app
	}

	return nil
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", er.Code, er.Message)
}

func (er ErrorResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(er.HTTPStatus)
	_ = Respond(w, &er)
}

func Respond(w http.ResponseWriter, data any) error {
	w.Header().Add("Content-Type", "application/json")
	dataStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataStr)
	return err
}

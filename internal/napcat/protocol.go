package napcat

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type PayloadType string

const (
	PaylaodRequest  PayloadType = "request"
	PayloadResponse PayloadType = "response"
	PayloadEvent    PayloadType = "event"
)

type Payload interface {
	PayloadType() PayloadType
}

type RequestType string

const (
	// OneBot11
	SendPrivateMsg       RequestType = "send_private_msg"
	SendGroupMsg         RequestType = "send_group_msg"
	SendMsg              RequestType = "send_msg"
	DeleteMsg            RequestType = "delete_msg"
	GetMsg               RequestType = "get_msg"
	GetForwardMsg        RequestType = "get_forward_msg"
	SendLike             RequestType = "send_like"
	SetGroupKick         RequestType = "set_group_kick"
	SetGroupBan          RequestType = "set_group_ban"
	SetGroupAnonymousBan RequestType = "set_group_anonymous_ban"
	SetGroupWholeBan     RequestType = "set_group_whole_ban"
	SetGroupAdmin        RequestType = "set_group_admin"
	SetGroupAnonymous    RequestType = "set_group_anonymous"
	SetGroupCard         RequestType = "set_group_card"
	SetGroupName         RequestType = "set_group_name"
	SetGroupLeave        RequestType = "set_group_leave"
	SetGroupSpecialTitle RequestType = "set_group_special_title"
	SetFriendAddRequest  RequestType = "set_friend_add_request"
	SetGroupAddRequest   RequestType = "set_group_add_request"
	GetLoginInfo         RequestType = "get_login_info"
	GetStrangerInfo      RequestType = "get_stranger_info"
	GetFriendList        RequestType = "get_friend_list"
	GetGroupInfo         RequestType = "get_group_info"
	GetGroupList         RequestType = "get_group_list"
	GetGroupMemberInfo   RequestType = "get_group_member_info"
	GetGroupMemberList   RequestType = "get_group_member_list"
	GetGroupHonorInfo    RequestType = "get_group_honor_info"
	GetCookies           RequestType = "get_cookies"
	GetCSRFToken         RequestType = "get_csrf_token"
	GetCredentials       RequestType = "get_credentials"
	GetRecord            RequestType = "get_record"
	GetImage             RequestType = "get_image"
	CanSendImage         RequestType = "can_send_image"
	CanSendRecord        RequestType = "can_send_record"
	GetStatus            RequestType = "get_status"
	GetVersionInfo       RequestType = "get_version_info"
	SetRestart           RequestType = "set_restart"
	CleanCache           RequestType = "clean_cache"

	// go-cqhttp
	MarkMsgAsRead         RequestType = "mark_msg_as_read"
	SendGroupForwardMsg   RequestType = "send_group_forward_msg"
	SendPrivateForwardMsg RequestType = "send_private_forward_msg"
	GetGroupMsgHistory    RequestType = "get_group_msg_history"
	UploadGroupFile       RequestType = "upload_group_file"
	DownloadFile          RequestType = "download_file"

	// NapCat
	GetFile                RequestType = "get_file"
	ForwardFriendSingleMsg RequestType = "forward_friend_single_msg"
	ForwardGroupSingleMsg  RequestType = "forward_group_single_msg"
	SendForwardMsg         RequestType = "send_forward_msg"
	MarkPrivateMsgAsRead   RequestType = "mark_private_msg_as_read"
	MarkGroupMsgAsRead     RequestType = "mark_private_msg_as_read"
	GetFriendMsgHistory    RequestType = "get_friend_msg_history"
)

type Request struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
	Echo   string                 `json:"echo,omitempty"`
}

func (r *Request) PayloadType() PayloadType {
	return PaylaodRequest
}

func NewGetLoginInfoRequest() *Request {
	return &Request{Action: string(GetLoginInfo)}
}

func NewGetUserInfoRequest(userID int64) *Request {
	return &Request{
		Action: string(GetStrangerInfo),
		Params: map[string]interface{}{
			"user_id": userID,
		},
	}
}

func NewGetGroupInfoRequest(groupID int64) *Request {
	return &Request{
		Action: string(GetGroupInfo),
		Params: map[string]interface{}{
			"group_id": groupID,
		},
	}
}

func NewGetFriendListRequest() *Request {
	return &Request{Action: string(GetFriendList)}
}

func NewGetGroupListRequest() *Request {
	return &Request{Action: string(GetGroupList)}
}

func NewGetGroupMemberListRequest(groupID int64) *Request {
	return &Request{
		Action: string(GetGroupMemberList),
		Params: map[string]interface{}{
			"group_id": groupID,
		},
	}
}

func NewGetGroupMemberInfoRequest(groupID, userID int64) *Request {
	return &Request{
		Action: string(GetGroupMemberInfo),
		Params: map[string]interface{}{
			"group_id": groupID,
			"user_id":  userID,
		},
	}
}

func NewGetRecordRequest(fileID string) *Request {
	return &Request{
		Action: string(GetRecord),
		Params: map[string]interface{}{
			"file":       fileID,
			"out_format": "ogg",
		},
	}
}

func NewGetImageRequest(fileID string) *Request {
	return &Request{
		Action: string(GetImage),
		Params: map[string]interface{}{
			"file_id": fileID,
		},
	}
}

func NewGetFileRequest(fileID string) *Request {
	return &Request{
		Action: string(GetFile),
		Params: map[string]interface{}{
			"file_id": fileID,
		},
	}
}

func NewGetForwardMsgRequest(msgID string) *Request {
	return &Request{
		Action: string(GetForwardMsg),
		Params: map[string]interface{}{
			"message_id": msgID,
		},
	}
}

func NewPrivateMsgRequest(userID int64, segments []ISegment) *Request {
	return &Request{
		Action: string(SendMsg),
		Params: map[string]interface{}{
			"message_type": "private",
			"user_id":      userID,
			"message":      segments,
		},
	}
}

func NewGroupMsgRequest(groupID int64, segments []ISegment) *Request {
	return &Request{
		Action: string(SendMsg),
		Params: map[string]interface{}{
			"message_type": "group",
			"group_id":     groupID,
			"message":      segments,
		},
	}
}

func NewDeleteMsgRequest(messageID int64) *Request {
	return &Request{
		Action: string(DeleteMsg),
		Params: map[string]interface{}{
			"message_id": messageID,
		},
	}
}

type Response struct {
	Status  string `json:"status"`
	Retcode int32  `json:"retcode"`
	Data    any    `json:"params,omitempty"`
	Echo    string `json:"echo,omitempty"`
}

func (r *Response) PayloadType() PayloadType {
	return PayloadResponse
}

type UserInfo struct {
	ID       int64  `json:"user_id" mapstructure:"user_id"`
	Nickname string `json:"nickname,omitempty" mapstructure:"nickname,omitempty"`
	Remark   string `json:"remark,omitempty" mapstructure:"remark,omitempty"`
}

type GroupInfo struct {
	ID   int64  `json:"group_id" mapstructure:"group_id"`
	Name string `json:"group_name,omitempty" mapstructure:"group_name,omitempty"`
}

type MemberInfo struct {
	UserID   int64  `json:"user_id" mapstructure:"user_id"`
	GroupID  int64  `json:"group_id" mapstructure:"group_id"`
	Nickname string `json:"nickname,omitempty" mapstructure:"nickname,omitempty"`
	Card     string `json:"card,omitempty" mapstructure:"card,omitempty"`
}

type FileInfo struct {
	ID       string `json:"id,omitempty" mapstructure:"id,omitempty"`
	Name     string `json:"name,omitempty" mapstructure:"name,omitempty"`
	File     string `json:"file,omitempty" mapstructure:"file,omitempty"`
	FileName string `json:"file_name,omitempty" mapstructure:"file_name,omitempty"`
	URL      string `json:"url" mapstructure:"url"`
	Base64   string `json:"base64,omitempty" mapstructure:"base64,omitempty"`
}

type EventType string

const (
	MessagePrivate             EventType = "message_private"
	MessageGroup               EventType = "message_group"
	MetaLifecycle              EventType = "meta_lifecycle"
	MetaHeartbeat              EventType = "meta_heartbeat"
	NoticeOfflineFile          EventType = "notice_offline_file"
	NoticeGroupUpload          EventType = "notice_group_upload"
	NoticeGroupAdminSet        EventType = "notice_group_admin_set"
	NoticeGroupAdminUnset      EventType = "notice_group_admin_unset"
	NoticeGroupDecreaseleave   EventType = "notice_group_decrease_leave"
	NoticeGroupDecreaseKick    EventType = "notice_group_decrease_kick"
	NoticeGroupDecreaseKickMe  EventType = "notice_group_decrease_kick_me"
	NoticeGroupIncreaseApprove EventType = "notice_group_increase_approve"
	NoticeGroupIncreaseInvite  EventType = "notice_group_increase_invite"
	NoticeGroupBanBan          EventType = "notice_group_ban_ban"
	NoticeGroupBanLiftBan      EventType = "notice_group_ban_lift_ban"
	NoticeFriendAdd            EventType = "notice_friend_add"
	NoticeGroupRecall          EventType = "notice_group_recall"
	NoticeFriendRecall         EventType = "notice_friend_recall"
	NoticeNotifyPoke           EventType = "notice_notify_poke"
	NoticeNotifyLuckyKing      EventType = "notice_notify_lucky_king"
	NoticeNotifyHonnor         EventType = "notice_notify_honnor"
	RequestFriend              EventType = "request_friend"
	RequestGroupAdd            EventType = "request_group_add"
	RequestGroupInvite         EventType = "request_group_invite"
	EventUnknown               EventType = "event_unknown"
)

type IEvent interface {
	EventType() EventType
}

type Event struct {
	Time     int64  `json:"time" mapstructure:"time"`
	SelfID   int64  `json:"self_id" mapstructure:"self_id"`
	PostType string `json:"post_type" mapstructure:"post_type"`
}

func (e *Event) PayloadType() PayloadType {
	return PayloadEvent
}

func (e *Event) EventType() EventType {
	return EventUnknown
}

type Message struct {
	Event       `mapstructure:",squash"`
	MessageType string    `json:"message_type" mapstructure:"message_type"`
	SubType     string    `json:"sub_type" mapstructure:"sub_type"`
	MessageID   int32     `json:"message_id" mapstructure:"message_id"`
	GroupID     int64     `json:"group_id,omitempty" mapstructure:"group_id,omitempty"`
	UserID      int64     `json:"user_id" mapstructure:"user_id"`
	TargetID    int64     `json:"target_id,omitempty" mapstructure:"target_id,omitempty"`
	Anonymous   Anonymous `json:"anonymous,omitempty" mapstructure:"anonymous,omitempty"`
	Message     any       `json:"message" mapstructure:"message"`
	Font        int32     `json:"font" mapstructure:"font"`
	Sender      Sender    `json:"sender" mapstructure:"sender"`
}

func (m *Message) EventType() EventType {
	if m.MessageType == "private" {
		return MessagePrivate
	}
	return MessageGroup
}

type Anonymous struct {
	ID   int64  `json:"id" mapstructure:"id"`
	Name string `json:"name" mapstructure:"name"`
	Flag string `json:"flag,omitempty" mapstructure:"flag,omitempty"`
}

type Sender struct {
	UserID   int64  `json:"user_id" mapstructure:"user_id"`
	Nickname string `json:"nickname,omitempty" mapstructure:"nickname,omitempty"`
	Card     string `json:"card,omitempty" mapstructure:"card,omitempty"`
	Sex      string `json:"sex,omitempty" mapstructure:"sex,omitempty"`
	Age      int32  `json:"age,omitempty" mapstructure:"age,omitempty"`
	Area     string `json:"area,omitempty" mapstructure:"area,omitempty"`
	Level    int32  `json:"level,omitempty" mapstructure:"level,omitempty"`
	Role     string `json:"role,omitempty" mapstructure:"role,omitempty"`
	Title    string `json:"title,omitempty" mapstructure:"title,omitempty"`
}

type LifeCycle struct {
	Event         `mapstructure:",squash"`
	MetaEventType string `json:"meta_event_type" mapstructure:"meta_event_type"`
	SubType       string `json:"sub_type" mapstructure:"sub_type"`
}

func (lc *LifeCycle) EventType() EventType {
	return MetaLifecycle
}

type Status struct {
	Online bool `json:"online" mapstructure:"online"`
	Good   bool `json:"good" mapstructure:"good"`
}

type Heartbeat struct {
	Event         `mapstructure:",squash"`
	MetaEventType string `json:"meta_event_type" mapstructure:"meta_event_type"`
	Status        Status `json:"status" mapstructure:"status"`
	Interval      int64  `json:"interval" mapstructure:"interval"`
}

func (h *Heartbeat) EventType() EventType {
	return MetaHeartbeat
}

type GroupRecall struct {
	Event      `mapstructure:",squash"`
	NoticeType string `json:"notice_type" mapstructure:"notice_type"`
	GroupID    int64  `json:"group_id" mapstructure:"group_id"`
	UserID     int64  `json:"user_id" mapstructure:"user_id"`
	OperatorID int64  `json:"operator_id" mapstructure:"operator_id"`
	MessageID  int64  `json:"message_id" mapstructure:"message_id"`
}

func (g *GroupRecall) EventType() EventType {
	return NoticeGroupRecall
}

type FriendRecall struct {
	Event      `mapstructure:",squash"`
	NoticeType string `json:"notice_type" mapstructure:"notice_type"`
	UserID     int64  `json:"user_id" mapstructure:"user_id"`
	MessageID  int64  `json:"message_id" mapstructure:"message_id"`
}

func (g *FriendRecall) EventType() EventType {
	return NoticeFriendRecall
}

type SegmentType string

const (
	Text       SegmentType = "text"
	Face       SegmentType = "face"
	MarketFace SegmentType = "mface"
	Image      SegmentType = "image"
	Record     SegmentType = "record"
	Video      SegmentType = "video"
	File       SegmentType = "file"
	At         SegmentType = "at"
	Share      SegmentType = "share"
	Contact    SegmentType = "contact"
	Location   SegmentType = "location"
	Music      SegmentType = "music"
	LightAPP   SegmentType = "lightapp"
	Reply      SegmentType = "reply"
	Forward    SegmentType = "forward"
	Node       SegmentType = "node"
	XML        SegmentType = "xml"
	JSON       SegmentType = "json"
)

type ISegment interface {
	SegmentType() SegmentType
}

type Segment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func (s *Segment) SegmentType() SegmentType {
	return SegmentType(s.Type)
}

type TextSegment struct {
	Segment `mapstructure:",squash"`
}

type FaceSegment struct {
	Segment `mapstructure:",squash"`
}

type MarketFaceSegment struct {
	Segment `mapstructure:",squash"`
}

type ImageSegment struct {
	Segment `mapstructure:",squash"`
}

type RecordSegment struct {
	Segment `mapstructure:",squash"`
}

type VideoSegment struct {
	Segment `mapstructure:",squash"`
}

type FileSegment struct {
	Segment `mapstructure:",squash"`
}

type AtSegment struct {
	Segment `mapstructure:",squash"`
}

type ShareSegment struct {
	Segment `mapstructure:",squash"`
}

type ContactSegment struct {
	Segment `mapstructure:",squash"`
}

type LocationSegment struct {
	Segment `mapstructure:",squash"`
}

type MusicSegment struct {
	Segment `mapstructure:",squash"`
}

type LightAPPSegment struct {
	Segment `mapstructure:",squash"`
}

type ReplySegment struct {
	Segment `mapstructure:",squash"`
}

type ForwardSegment struct {
	Segment `mapstructure:",squash"`
}

type NodeSegment struct {
	Segment `mapstructure:",squash"`
}

type XMLSegment struct {
	Segment `mapstructure:",squash"`
}

type JSONSegment struct {
	Segment `mapstructure:",squash"`
}

func (s *TextSegment) Content() string {
	return s.Data["text"].(string)
}

func (s *FaceSegment) ID() string {
	return s.Data["id"].(string)
}

func (s *MarketFaceSegment) Content() string {
	return s.Data["summary"].(string)
}

func (s *MarketFaceSegment) URL() string {
	return s.Data["url"].(string)
}

func (s *MarketFaceSegment) File() string {
	return s.Data["file"].(string)
}

func (s *ImageSegment) File() string {
	return s.Data["file"].(string)
}

func (s *ImageSegment) URL() string {
	return s.Data["url"].(string)
}

func (s *ImageSegment) IsSticker() bool {
	_, exists := s.Data["emoji_id"]
	return exists
}

func (s *RecordSegment) File() string {
	return s.Data["file"].(string)
}

func (s *VideoSegment) URL() string {
	return s.Data["url"].(string)
}

func (s *VideoSegment) File() string {
	return s.Data["file"].(string)
}

func (s *FileSegment) File() string {
	return s.Data["file"].(string)
}

func (s *AtSegment) Target() string {
	return s.Data["qq"].(string)
}

func (s *ShareSegment) URL() string {
	return s.Data["url"].(string)
}

func (s *ShareSegment) Title() string {
	return s.Data["title"].(string)
}

func (s *ShareSegment) Content() string {
	return s.Data["content"].(string)
}

func (s *ShareSegment) Image() string {
	return s.Data["image"].(string)
}

func (s *LocationSegment) Latitude() float64 {
	return s.Data["lat"].(float64)
}

func (s *LocationSegment) Longitude() float64 {
	return s.Data["lon"].(float64)
}

func (s *LocationSegment) Title() string {
	return s.Data["title"].(string)
}

func (s *LocationSegment) Content() string {
	return s.Data["content"].(string)
}

func (s *ReplySegment) ID() string {
	return s.Data["id"].(string)
}

func (s *ForwardSegment) ID() string {
	return s.Data["id"].(string)
}

func (s *NodeSegment) ID() string {
	return s.Data["id"].(string)
}

func (s *XMLSegment) Content() string {
	return s.Data["data"].(string)
}

func (s *JSONSegment) Content() string {
	return s.Data["data"].(string)
}

func NewText(content string) *TextSegment {
	return &TextSegment{
		Segment{
			Type: string(Text),
			Data: map[string]interface{}{
				"text": content,
			},
		},
	}
}

func NewFace(id string) *FaceSegment {
	return &FaceSegment{
		Segment{
			Type: string(Face),
			Data: map[string]interface{}{
				"id": id,
			},
		},
	}
}

func NewImage(file, name string) *ImageSegment {
	return &ImageSegment{
		Segment: Segment{
			Type: string(Image),
			Data: map[string]interface{}{
				"file": file,
				"name": name,
			},
		},
	}
}

func NewRecord(file, name string) *RecordSegment {
	return &RecordSegment{
		Segment{
			Type: string(Record),
			Data: map[string]interface{}{
				"file": file,
				"name": name,
			},
		},
	}
}

func NewVideo(file, name string) *VideoSegment {
	return &VideoSegment{
		Segment{
			Type: string(Video),
			Data: map[string]interface{}{
				"file": file,
				"name": name,
			},
		},
	}
}

func NewFile(file, name string) *FileSegment {
	return &FileSegment{
		Segment{
			Type: string(File),
			Data: map[string]interface{}{
				"file": file,
				"name": name,
			},
		},
	}
}

func NewAt(target string) *AtSegment {
	return &AtSegment{
		Segment{
			Type: string(At),
			Data: map[string]interface{}{
				"qq": target,
			},
		},
	}
}

func NewReply(id string) *ReplySegment {
	return &ReplySegment{
		Segment{
			Type: string(Reply),
			Data: map[string]interface{}{
				"id": id,
			},
		},
	}
}

func NewJSON(content string) *NodeSegment {
	return &NodeSegment{
		Segment{
			Type: string(JSON),
			Data: map[string]interface{}{
				"data": content,
			},
		},
	}
}

func UnmarshalPayload(m map[string]interface{}) (Payload, error) {
	if postType, ok := m["post_type"]; ok {
		switch postType {
		case "message":
			return unmarshalMessage(m)
		case "message_sent":
			return unmarshalMessage(m)
		case "meta_event":
			return unmarshalMeta(m)
		case "notice":
			return unmarshalNotice(m)
		case "request":
			return unmarshalEvent(m)
		}
		return nil, fmt.Errorf("event %s is not supported", postType)
	} else if _, ok := m["retcode"]; ok {
		return unmarshalResponse(m)
	} else if _, ok := m["action"]; ok {
		return unmarshalRequest(m)
	}

	return nil, errors.New("payload type is not supported")
}

func unmarshalMeta(m map[string]interface{}) (Payload, error) {
	switch m["meta_event_type"] {
	case "lifecycle":
		var event LifeCycle
		err := mapstructure.WeakDecode(m, &event)
		return &event, err
	case "heartbeat":
		var event Heartbeat
		err := mapstructure.WeakDecode(m, &event)
		return &event, err
	}

	return unmarshalEvent(m)
}

func unmarshalNotice(m map[string]interface{}) (Payload, error) {
	switch m["notice_type"] {
	case "group_recall":
		var event GroupRecall
		err := mapstructure.WeakDecode(m, &event)
		return &event, err
	case "friend_recall":
		var event FriendRecall
		err := mapstructure.WeakDecode(m, &event)
		return &event, err
	}

	return unmarshalEvent(m)
}

func unmarshalEvent(m map[string]interface{}) (Payload, error) {
	var event Event
	err := mapstructure.WeakDecode(m, &event)
	return &event, err
}

func unmarshalRequest(m map[string]interface{}) (Payload, error) {
	var event Request
	err := mapstructure.WeakDecode(m, &event)
	return &event, err
}

func unmarshalResponse(m map[string]interface{}) (Payload, error) {
	var event Response
	err := mapstructure.WeakDecode(m, &event)
	return &event, err
}

func unmarshalMessage(m map[string]interface{}) (Payload, error) {
	var event Message
	if err := mapstructure.WeakDecode(m, &event); err != nil {
		return nil, err
	}

	if m["message"] != nil {
		event.Message = generateSegments(m["message"].([]interface{}))
	} else if m["content"] != nil {
		event.Message = generateSegments(m["content"].([]interface{}))
	}
	return &event, nil
}

func generateSegments(d []interface{}) []ISegment {
	segments := []ISegment{}

	for _, s := range d {
		switch s.(map[string]interface{})["type"].(string) {
		case string(Text):
			var segment TextSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Face):
			var segment FaceSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(MarketFace):
			var segment MarketFaceSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Image):
			var segment ImageSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Record):
			var segment RecordSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Video):
			var segment VideoSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(File):
			var segment FileSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(At):
			var segment AtSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Share):
			var segment ShareSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Location):
			var segment LocationSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Reply):
			var segment ReplySegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Forward):
			var segment ForwardSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(Node):
			var segment NodeSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(XML):
			var segment XMLSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		case string(JSON):
			var segment JSONSegment
			mapstructure.WeakDecode(s, &segment)
			segments = append(segments, &segment)
		}
	}

	return segments
}

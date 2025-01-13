package napcat

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/duo/matrix-octopus-napcat/internal/common"

	"github.com/gabriel-vasile/mimetype"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
)

func (c *Client) handleResponse(resp *Response) {
	c.websocketRequestsLock.RLock()
	respChan, ok := c.websocketRequests[resp.Echo]
	c.websocketRequestsLock.RUnlock()
	if ok {
		select {
		case respChan <- resp:
		default:
			c.log.Warn().Msgf("Failed to handle response to %s: channel didn't accept response", resp.Echo)
		}
	} else {
		c.log.Warn().Msgf("Dropping response to %s: unknown request ID", resp.Echo)
	}
}

func (c *Client) handleEvent(evt IEvent, transmitFunc func(*common.Packet) error) {
	key := c.getEventKey(evt)
	c.mutex.LockKey(key)
	defer c.mutex.UnlockKey(key)

	var msg *common.Message

	switch evt.EventType() {
	case MessagePrivate:
		msg = c.handlePrivateMessage(evt.(*Message))
	case MessageGroup:
		msg = c.handleGroupMessage(evt.(*Message))
	case NoticeGroupRecall:
		msg = c.handleGroupRecall(evt.(*GroupRecall))
	case NoticeFriendRecall:
		msg = c.handleFriendRecall(evt.(*FriendRecall))
	case MetaLifecycle:
		// TODO:
	case MetaHeartbeat:
		transmitFunc(&common.Packet{
			ID:      0,
			MXID:    c.mxid,
			Type:    common.PktNotice,
			Payload: c.handleHeartbeat(evt.(*Heartbeat)),
		})
	}

	if msg != nil {
		transmitFunc(&common.Packet{
			ID:      0,
			MXID:    c.mxid,
			Type:    common.PktMessage,
			Payload: msg,
		})
	}
}

func (c *Client) handlePrivateMessage(evt *Message) *common.Message {
	chatID := evt.Sender.UserID
	if evt.PostType == "message_sent" { // sent by self
		chatID = evt.TargetID
	}

	message := &common.Message{
		ID:        fmt.Sprint(evt.MessageID),
		Timestamp: evt.Time * 1000,
		From: common.User{
			ID: fmt.Sprint(evt.Sender.UserID),
		},
		Chat: common.Chat{
			ID:   fmt.Sprint(chatID),
			Type: common.ChatPrivate,
		},
		Mentions: []string{},
	}

	c.populateMessage(message, evt.Message.([]ISegment))

	return message
}

func (c *Client) handleGroupMessage(evt *Message) *common.Message {
	message := &common.Message{
		ID:        fmt.Sprint(evt.MessageID),
		Timestamp: evt.Time * 1000,
		From: common.User{
			ID: fmt.Sprint(evt.Sender.UserID),
		},
		Chat: common.Chat{
			ID:   fmt.Sprint(evt.GroupID),
			Type: common.ChatGroup,
		},
		Mentions: []string{},
	}

	c.populateMessage(message, evt.Message.([]ISegment))

	return message
}

func (c *Client) populateMessage(msg *common.Message, segments []ISegment) {
	msg.Type = common.MsgText

	var summary []string

	images := []*common.BlobData{}
	for _, s := range segments {
		switch v := s.(type) {
		case *TextSegment:
			summary = append(summary, v.Content())
		case *FaceSegment:
			summary = append(summary, fmt.Sprintf("/[Face%s]", v.ID()))
		case *MarketFaceSegment:
			summary = append(summary, v.Content())
			if blob, err := c.getMedia(v); err != nil {
				c.log.Warn().Err(err).Msg("Failed to download market face")
			} else {
				msg.Type = common.MsgSticker
				msg.Attachment = []*common.BlobData{blob}
			}
		case *AtSegment:
			target := v.Target()
			if target == "all" {
				target = "room" // Matrix's mention all
			}
			summary = append(summary, fmt.Sprintf("@%s ", target))
			msg.Mentions = append(msg.Mentions, target)
		case *ImageSegment:
			if blob, err := c.getMedia(v); err != nil {
				c.log.Warn().Err(err).Msg("Failed to download image")
				summary = append(summary, "[图片下载失败]")
			} else {
				images = append(images, blob)
				summary = append(summary, "[图片]")
			}
		case *VideoSegment:
			if blob, err := c.getMedia(v); err != nil {
				c.log.Warn().Err(err).Msg("Failed to download video")
				msg.Content = "[视频下载失败]"
			} else {
				msg.Type = common.MsgVideo
				msg.Attachment = []*common.BlobData{blob}
			}
		case *FileSegment:
			if blob, err := c.getMedia(v); err != nil {
				c.log.Warn().Err(err).Msg("Failed to download file")
				msg.Content = "[文件下载失败]"
			} else {
				msg.Type = common.MsgFile
				msg.Attachment = []*common.BlobData{blob}
			}
		case *RecordSegment:
			if blob, err := c.getMedia(v); err != nil {
				c.log.Warn().Err(err).Msg("Failed to download record")
				msg.Content = "[语音下载失败]"
			} else {
				msg.Type = common.MsgAudio
				msg.Attachment = []*common.BlobData{blob}
			}
		case *ReplySegment:
			msg.Reply = &common.ReplyInfo{
				ID: v.ID(),
			}
		case *ForwardSegment:
			// TODO:
			msg.Content = "[聊天记录]"
		case *JSONSegment:
			content := v.Content()
			view := gjson.Get(content, "view").String()
			if view == "LocationShare" {
				name := gjson.Get(content, "meta.*.name").String()
				address := gjson.Get(content, "meta.*.address").String()
				latitude := gjson.Get(content, "meta.*.lat").Float()
				longitude := gjson.Get(content, "meta.*.lng").Float()
				msg.Type = common.MsgLocation
				msg.Attachment = &common.LocationData{
					Name:      name,
					Address:   address,
					Longitude: longitude,
					Latitude:  latitude,
				}
			} else {
				if url := gjson.Get(content, "meta.*.qqdocurl").String(); len(url) > 0 {
					desc := gjson.Get(content, "meta.*.desc").String()
					prompt := gjson.Get(content, "prompt").String()
					msg.Type = common.MsgApp
					msg.Attachment = &common.AppData{
						Title:       prompt,
						Description: desc,
						URL:         url,
					}
				} else if jumpUrl := gjson.Get(content, "meta.*.jumpUrl").String(); len(jumpUrl) > 0 {
					desc := gjson.Get(content, "meta.*.desc").String()
					prompt := gjson.Get(content, "prompt").String()
					msg.Type = common.MsgApp
					msg.Attachment = &common.AppData{
						Title:       prompt,
						Description: desc,
						URL:         jumpUrl,
					}
				}
			}
		default:
			summary = append(summary, fmt.Sprintf("[%v]", s.SegmentType()))
		}
	}

	if len(summary) > 0 {
		if len(summary) == 1 && segments[0].SegmentType() == Image {
			msg.Type = common.MsgImage
			msg.Attachment = images

			if segments[0].(*ImageSegment).IsSticker() && len(images) == 1 {
				msg.Type = common.MsgSticker
			}
		} else {
			msg.Content = strings.Join(summary, "")

			if len(images) > 0 {
				msg.Type = common.MsgImage
				msg.Attachment = images
			}
		}
	}
}

func (c *Client) handleFriendRecall(evt *FriendRecall) *common.Message {
	return &common.Message{
		ID:        fmt.Sprint(evt.MessageID),
		Timestamp: evt.Time * 1000,
		From: common.User{
			ID: fmt.Sprint(evt.UserID),
		},
		Chat: common.Chat{
			ID:   fmt.Sprint(evt.UserID),
			Type: common.ChatPrivate,
		},
		Type: common.MsgRevoke,
	}
}

func (c *Client) handleGroupRecall(evt *GroupRecall) *common.Message {
	return &common.Message{
		ID:        fmt.Sprint(evt.MessageID),
		Timestamp: evt.Time * 1000,
		Content:   fmt.Sprint(evt.OperatorID),
		From: common.User{
			ID: fmt.Sprint(evt.UserID),
		},
		Chat: common.Chat{
			ID:   fmt.Sprint(evt.GroupID),
			Type: common.ChatGroup,
		},
		Type: common.MsgRevoke,
	}
}

func (c *Client) handleHeartbeat(evt *Heartbeat) *common.Notice {
	c.uin = evt.SelfID

	return &common.Notice{
		ClientID:  fmt.Sprint(evt.SelfID),
		Timestamp: evt.Time * 1000,
		Type:      common.NoticeHeartbeat,
		Data: &common.Heartbeat{
			IsLoggedIn: evt.Status.Online,
			Interval:   uint32(evt.Interval),
		},
	}
}

func (c *Client) getEventKey(evt IEvent) string {
	switch evt.EventType() {
	case MessagePrivate:
		m := evt.(*Message)
		targetID := m.Sender.UserID
		if m.PostType == "message_sent" { // sent by self
			targetID = m.TargetID
		}
		return fmt.Sprint(targetID)
	case MessageGroup:
		return fmt.Sprint(evt.(*Message).GroupID)
	case NoticeGroupRecall:
		return fmt.Sprint(evt.(*GroupRecall).GroupID)
	case NoticeFriendRecall:
		return fmt.Sprint(evt.(*FriendRecall).UserID)
	}

	return ""
}

func (c *Client) getMedia(segment ISegment) (*common.BlobData, error) {
	var request *Request
	var url string

	switch v := segment.(type) {
	case *ImageSegment:
		request = NewGetImageRequest(v.File())
		url = v.URL()
	case *MarketFaceSegment:
		request = NewGetImageRequest(v.File())
		url = v.URL()
	case *VideoSegment:
		request = NewGetFileRequest(v.File())
		url = v.URL()
	case *FileSegment:
		request = NewGetFileRequest(v.File())
	case *RecordSegment:
		request = NewGetRecordRequest(v.File())
	default:
		return nil, fmt.Errorf("media type not supported: %+v", v.SegmentType())
	}

	if segment.SegmentType() == MarketFace || segment.SegmentType() == Video ||
		(segment.SegmentType() == Image && segment.(*ImageSegment).IsSticker()) {
		if strings.HasPrefix(url, "http") {
			if blob, err := Download(url); err == nil {
				return blob, nil
			}
		}
	}

	if resp, err := c.request(request); err == nil {
		var f FileInfo
		if err := mapstructure.WeakDecode(resp, &f); err != nil {
			return nil, err
		}

		if f.Base64 != "" {
			if data, err := base64.StdEncoding.DecodeString(f.Base64); err == nil {
				return &common.BlobData{
					Name:   f.FileName,
					Mime:   mimetype.Detect(data).String(),
					Binary: data,
				}, nil
			}
		}
	}

	if strings.HasPrefix(url, "http") {
		if blob, err := Download(url); err == nil {
			return blob, nil
		}
	}

	return nil, fmt.Errorf("Failed to download media: %+v", segment)
}

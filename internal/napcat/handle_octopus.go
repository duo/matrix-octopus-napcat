package napcat

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/duo/matrix-octopus-napcat/internal/common"

	"github.com/go-cmd/cmd"
	"github.com/mitchellh/mapstructure"
)

func (c *Client) Connect() error {
	c.instanceLock.Lock()
	defer c.instanceLock.Unlock()

	c.log.Info().Msg("Connect to a NapCat instance")

	if c.instance != nil && !c.instance.Status().Complete {
		c.log.Info().Msg("NapCat instance already exists")
		return nil
	}

	// Overwrite WebUI configuration
	if err := c.overwriteWebUIConfig(); err != nil {
		return err
	}

	// Start the instance
	parts := strings.Fields(c.command)
	instance := cmd.NewCmd(parts[0], parts[1:]...)
	instance.Start()

	// Wait for the instance to start up
	started := false
	ctx, cancel := context.WithTimeout(context.Background(), c.initTimeout)
	defer cancel()
	for {
		if instance.Status().PID != 0 {
			started = true
			break
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			break
		}
	}

	if !started {
		return fmt.Errorf("failed to start the client")
	}

	c.instance = instance

	return nil
}

func (c *Client) Disconnect() {
	c.instanceLock.Lock()
	defer c.instanceLock.Unlock()

	c.log.Info().Msg("Disconnect from the NapCat instance")

	if c.instance != nil {
		c.instance.Stop()
		c.instance = nil
	}

	// Cleanup configuration files
	os.Remove(filepath.Join(c.path, fmt.Sprintf("napcat_%d.json", c.uin)))
	os.Remove(filepath.Join(c.path, fmt.Sprintf("onebot11_%d.json", c.uin)))
}

func (c *Client) Login(step *common.LoginStep) (*common.LoginStep, error) {
	c.instanceLock.Lock()
	defer c.instanceLock.Unlock()

	switch step.Type {
	case common.LoginStepTypeDisplayAndWait:
		{
			if step.DisplayAndWaitParams.Type != common.LoginDisplayTypeQR {
				return nil, fmt.Errorf("display type %s is not supported", step.DisplayAndWaitParams.Type)
			}

			if resp, err := c.checkLogin(); err != nil {
				return nil, err
			} else {
				if resp.Data.IsLogin {
					c.updateOB11Config()
					return c.genCompleteStep()
				} else {
					return &common.LoginStep{
						StepID:       LoginStepQR,
						Type:         common.LoginStepTypeDisplayAndWait,
						Instructions: "Scan the QR code on your app to log in",
						DisplayAndWaitParams: &common.LoginDisplayAndWaitParams{
							Type: common.LoginDisplayTypeQR,
							Data: resp.Data.QRCodeURL,
						},
					}, nil
				}
			}
		}
	case common.LoginStepTypeWait:
		{
			if resp, err := c.checkLogin(); err != nil {
				return nil, err
			} else {
				if resp.Data.IsLogin {
					c.updateOB11Config()
					return c.genCompleteStep()
				} else {
					return &common.LoginStep{
						StepID: LoginStepWait,
						Type:   common.LoginStepTypeWait,
					}, nil
				}
			}
		}
	default:
		return nil, fmt.Errorf("step %s is not supported", step.Type)
	}
}

func (c *Client) Logout() (bool, error) {
	// TODO: logout from NapCat
	return true, nil
}

func (c *Client) GetLoginInfo() (*common.UserInfo, error) {
	if resp, err := c.getLoginInfo(); err != nil {
		return nil, err
	} else {
		return &common.UserInfo{
			ID:   resp.Data.Uin,
			Name: resp.Data.Nick,
		}, nil
	}
}

func (c *Client) GetUserInfo(id string) (*common.UserInfo, error) {
	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(NewGetUserInfoRequest(userID))
	if err != nil {
		return nil, err
	}

	var info UserInfo
	if err := mapstructure.WeakDecode(resp, &info); err != nil {
		return nil, err
	}

	return &common.UserInfo{
		ID:     fmt.Sprint(info.ID),
		Name:   info.Nickname,
		Alias:  info.Remark,
		Avatar: getUserAvatarURL(userID),
	}, nil
}

func (c *Client) GetGroupInfo(id string) (*common.GroupInfo, error) {
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(NewGetGroupInfoRequest(groupID))
	if err != nil {
		return nil, err
	}

	var info GroupInfo
	if err := mapstructure.WeakDecode(resp, &info); err != nil {
		return nil, err
	}

	members, err := c.GetGroupMemberList(id)
	if err != nil {
		return nil, err
	}

	groupInfo := &common.GroupInfo{
		ID:      fmt.Sprint(info.ID),
		Name:    info.Name,
		Avatar:  getGroupAvatarURL(groupID),
		Members: make([]string, 0, len(members)),
	}
	for _, member := range members {
		groupInfo.Members = append(groupInfo.Members, fmt.Sprint(member.ID))
	}

	return groupInfo, nil
}

func (c *Client) GetFriendList() ([]*common.UserInfo, error) {
	resp, err := c.request(NewGetFriendListRequest())
	if err != nil {
		return nil, err
	}

	var friends []*UserInfo
	if err := mapstructure.WeakDecode(resp, &friends); err != nil {
		return nil, err
	}

	friendList := make([]*common.UserInfo, 0, len(friends))
	for _, friend := range friends {
		friendList = append(friendList, &common.UserInfo{
			ID:     fmt.Sprint(friend.ID),
			Name:   friend.Nickname,
			Alias:  friend.Remark,
			Avatar: getUserAvatarURL(friend.ID),
		})
	}

	return friendList, nil
}

func (c *Client) GetGroupList() ([]*common.GroupInfo, error) {
	resp, err := c.request(NewGetGroupListRequest())
	if err != nil {
		return nil, err
	}

	var groups []*GroupInfo
	if err := mapstructure.WeakDecode(resp, &groups); err != nil {
		return nil, err
	}

	groupList := make([]*common.GroupInfo, 0, len(groups))
	for _, group := range groups {
		groupList = append(groupList, &common.GroupInfo{
			ID:     fmt.Sprint(group.ID),
			Name:   group.Name,
			Avatar: getGroupAvatarURL(group.ID),
		})
	}

	return groupList, nil
}

func (c *Client) GetGroupMemberList(id string) ([]*common.MemberInfo, error) {
	groupID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(NewGetGroupMemberListRequest(groupID))
	if err != nil {
		return nil, err
	}

	var members []*MemberInfo
	if err := mapstructure.WeakDecode(resp, &members); err != nil {
		return nil, err
	}

	memberList := make([]*common.MemberInfo, 0, len(members))
	for _, member := range members {
		memberList = append(memberList, &common.MemberInfo{
			ID:     fmt.Sprint(member.UserID),
			Name:   member.Nickname,
			Alias:  member.Card,
			Avatar: getUserAvatarURL(member.UserID),
		})
	}

	return memberList, nil
}

func (c *Client) GetGroupMemberInfo(gid string, uid string) (*common.MemberInfo, error) {
	groupID, err := strconv.ParseInt(gid, 10, 64)
	if err != nil {
		return nil, err
	}
	userID, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(NewGetGroupMemberInfoRequest(groupID, userID))
	if err != nil {
		return nil, err
	}

	var member *MemberInfo
	if err := mapstructure.WeakDecode(resp, &member); err != nil {
		return nil, err
	}

	return &common.MemberInfo{
		ID:     fmt.Sprint(member.UserID),
		Name:   member.Nickname,
		Alias:  member.Card,
		Avatar: getUserAvatarURL(member.UserID),
	}, nil
}

func (c *Client) SendMessage(msg *common.Message) (*common.Message, error) {
	chatID, err := strconv.ParseInt(msg.Chat.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	segments := []ISegment{}

	if msg.Reply != nil {
		segments = append(segments, NewReply(msg.Reply.ID))
	}

	switch msg.Type {
	case common.MsgText:
		segments = append(segments, c.genTextSegments(msg.Content, msg.Mentions)...)
	case common.MsgImage:
		images := msg.Attachment.([]*common.BlobData)
		for _, image := range images {
			binary := fmt.Sprintf("base64://%s", base64.StdEncoding.EncodeToString(image.Binary))
			segments = append(segments, NewImage(binary, image.Name))
		}
	case common.MsgSticker:
		blob := msg.Attachment.([]*common.BlobData)[0]
		binary := fmt.Sprintf("base64://%s", base64.StdEncoding.EncodeToString(blob.Binary))
		segments = append(segments, NewImage(binary, blob.Name))
	case common.MsgVideo:
		blob := msg.Attachment.([]*common.BlobData)[0]
		binary := fmt.Sprintf("base64://%s", base64.StdEncoding.EncodeToString(blob.Binary))
		segments = append(segments, NewVideo(binary, blob.Name))
	case common.MsgAudio:
		blob := msg.Attachment.([]*common.BlobData)[0]
		binary := fmt.Sprintf("base64://%s", base64.StdEncoding.EncodeToString(blob.Binary))
		segments = append(segments, NewRecord(binary, blob.Name))
	case common.MsgFile:
		blob := msg.Attachment.([]*common.BlobData)[0]
		binary := fmt.Sprintf("base64://%s", base64.StdEncoding.EncodeToString(blob.Binary))
		segments = append(segments, NewFile(binary, blob.Name))
	case common.MsgLocation:
		location := msg.Attachment.(*common.LocationData)
		locationJson := fmt.Sprintf(`
		{
			"app": "com.tencent.map",
			"desc": "地图",
			"view": "LocationShare",
			"ver": "0.0.0.1",
			"prompt": "[位置]%s",
			"from": 1,
			"meta": {
			  "Location.Search": {
				"id": "12250896297164027526",
				"name": "%s",
				"address": "%s",
				"lat": "%.5f",
				"lng": "%.5f",
				"from": "plusPanel"
			  }
			},
			"config": {
			  "forward": 1,
			  "autosize": 1,
			  "type": "card"
			}
		}
		`, location.Name, location.Name, location.Address, location.Latitude, location.Longitude)
		segments = append(segments, NewJSON(locationJson))
	default:
		return nil, fmt.Errorf("%s not support", msg.Type)
	}

	var request *Request
	if msg.Chat.Type == common.ChatPrivate {
		request = NewPrivateMsgRequest(chatID, segments)
	} else {
		request = NewGroupMsgRequest(chatID, segments)
	}

	if resp, err := c.request(request); err != nil {
		return nil, err
	} else {
		messageID := int64(resp.(map[string]interface{})["message_id"].(float64))
		return &common.Message{
			ID:        fmt.Sprint(messageID),
			ThreadID:  msg.ThreadID,
			Timestamp: time.Now().UnixMilli(),
			From:      msg.From,
			Chat:      msg.Chat,
			Type:      common.MsgSystem,
		}, nil
	}
}

func (c *Client) RevokeMessage(id string) error {
	messageID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	request := NewDeleteMsgRequest(messageID)

	_, err = c.request(request)

	return err
}

func (c *Client) genTextSegments(content string, mentions []string) []ISegment {
	if len(mentions) == 0 {
		return []ISegment{NewText(content)}
	}

	keywords := make([]string, len(mentions))
	for i, m := range mentions {
		keywords[i] = "@" + m
	}

	pattern := strings.Join(keywords, "|")
	re := regexp.MustCompile("(?:" + pattern + ")")

	parts := re.Split(content, -1)
	matches := re.FindAllString(content, -1)

	var splits []string
	for i := 0; i < len(parts); i++ {
		if parts[i] != "" {
			splits = append(splits, parts[i])
		}
		if i < len(matches) {
			splits = append(splits, matches[i])
		}
	}

	segments := []ISegment{}
	for _, s := range splits {
		if slices.Contains(keywords, s) {
			if s == "@room" {
				segments = append(segments, NewAt("all"))
			} else {
				segments = append(segments, NewAt(s[1:]))
			}
		} else {
			segments = append(segments, NewText(s))
		}
	}

	return segments
}

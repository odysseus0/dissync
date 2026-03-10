package discord

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type Channel struct {
	ID            string `json:"id"`
	GuildID       string `json:"guild_id"`
	Name          string `json:"name"`
	Type          int    `json:"type"`
	ParentID      string `json:"parent_id"`
	Position      int    `json:"position"`
	LastMessageID string `json:"last_message_id"`
}

// Discord channel types that can contain messages.
const (
	ChannelTypeGuildText         = 0
	ChannelTypeGuildAnnouncement = 5
	ChannelTypeAnnouncementThread = 10
	ChannelTypePublicThread      = 11
	ChannelTypePrivateThread     = 12
	ChannelTypeGuildForum        = 15
)

func (c Channel) HasMessages() bool {
	switch c.Type {
	case ChannelTypeGuildText, ChannelTypeGuildAnnouncement,
		ChannelTypeAnnouncementThread, ChannelTypePublicThread, ChannelTypePrivateThread:
		return true
	}
	return false
}

type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type MessageReference struct {
	MessageID string `json:"message_id"`
}

type Message struct {
	ID              string            `json:"id"`
	ChannelID       string            `json:"channel_id"`
	GuildID         string            `json:"guild_id"`
	Author          Author            `json:"author"`
	Content         string            `json:"content"`
	Type            int               `json:"type"`
	MessageReference *MessageReference `json:"message_reference"`
	EditedTimestamp  *string           `json:"edited_timestamp"`
	Timestamp       string            `json:"timestamp"`
}

// Thread search response (undocumented user-token endpoint).
type ThreadSearchResponse struct {
	Threads []Channel `json:"threads"`
	HasMore bool      `json:"has_more"`
	FirstID string    `json:"first_id"`
}

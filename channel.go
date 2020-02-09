package botserver

import (
	"github.com/sirupsen/logrus"
	"strings"
)

// Callback is
type Callback struct {
	// Time the message was created (unix time)
	CreatedAt uint64 `json:"created_at"`
	// The id of the group
	GroupID string `json:"group_id"`
	// UID of the message
	MessageID string `json:"id"`
	// UID of the sender
	SenderID string `json:"sender_id"`
	// Name of the sender (This is the nickname, and not static)
	SenderName string `json:"name"`
	// Type of the sender. I Have no idea what this means, I'll be honest.
	SenderType string `json:"user"`
	// I have no idea...
	SourceGUID string `json:"source_guid"`
	// Whether or not this is a system message.
	IsSystem bool `json:"system"`
	// The text of the message
	Text string `json:"text"`
	// The user id. I have no idea what that is
	UserID string `json:"user_id"`
}

// Channel to register to the groupme server.
// Register Hooks to a channel to add individual functions to the channel
// while maintaining modularity.
type Channel struct {
	// The IDs that the channel serves
	GroupIDs []string
	hooks    []hookContainer
	// name for debugging purposes
	Name             string
	// Server log, initialized when the channel is added to a server instance
	log *logrus.Logger
	rebuildIndexFunc func()
}

// Adds the hook into the channel. No duplicate checking. Returns the registered ID.
func (c *Channel) AddHook(h Hook) int32 {
	wrapper := wrapHook(h)
	c.hooks = append(c.hooks, wrapper)
	return wrapper.id
}

// BUG(EthanZeigler): Removal of a hook After it has been registered to a channel will leak memory
// Removes the hook of the given ID. ID is returned when the hook is attached.
func (c *Channel) RemoveHook(id int32) {
	for i := 0; i < len(c.hooks); i++ {
		if c.hooks[i].id == id {
			c.hooks = append(c.hooks[:i], c.hooks[i+1:]...)
			return
		}
	}
}

// Sends asynchronous events to each hook in the channel
func (c Channel) inputCallback(input Callback) {
	input.Text = strings.Trim(input.Text, " \n")
	c.log.WithFields(logrus.Fields{
		"sender": input.SenderName,
		"msg": input.Text,
	}).Debug("Sending callback to hooks")

	for _, wrapper := range c.hooks {
		c.log.WithField("hook", wrapper.hook.Name()).Debug("Alerting hook")
		if wrapper.hook.Action(input) {
			c.log.WithField("hook", wrapper.hook.Name()).Debug("Served by hook")
			break
		} else {
			c.log.WithField("hook", wrapper.hook.Name()).Debug("Not served by hook")
		}
	}
}

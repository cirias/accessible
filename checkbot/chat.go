package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cirias/accessible"
	"github.com/cirias/accessible/telbot"
)

type Chat struct {
	id        telbot.ChatId
	sb        *Subscriber
	subConfig struct {
		Name     string
		URL      string
		Duration time.Duration
	}
	update func(*telbot.Update) *telbot.MessageParams
}

func NewChat(client *accessible.Client) *Chat {
	c := &Chat{
		sb: NewSubscriber(client),
	}
	c.update = c.updateIdle
	return c
}

func (c *Chat) Handle(send telbot.SendMessageFunc, u *telbot.Update) {
	msg := c.update(u)

	if _, err := send(msg); err != nil {
		log.Printf("could not send: %s\n", err)
	}
}

func (c *Chat) updateIdle(u *telbot.Update) *telbot.MessageParams {
	var msg *telbot.MessageParams
	switch u.Message.Command() {
	case "sub":
		c.update = c.updateSubNew
		msg = c.textMsg("Name?")

	case "ls":
		// TODO
		msg = c.textMsg("TODO")

	case "log":
		arg := u.Message.CommandArguments()
		sub, ok := c.sb.Subscriptions().Load(arg)
		if !ok {
			msg = c.textMsg("could not found")
			break
		}
		msg = c.textMsg(fmt.Sprint(sub.(*Subscription).History()))

	case "rm":
		arg := u.Message.CommandArguments()
		if err := c.sb.Unsubscribe(arg); err != nil {
			msg = c.textMsg(fmt.Sprintf("could not unsubscribe: %s", err))
			break
		}
		msg = c.textMsg("Removed")
	}
	return msg
}

func (c *Chat) updateSubNew(u *telbot.Update) *telbot.MessageParams {
	if u.Message.Command() == "cancel" {
		c.update = c.updateIdle
		return c.textMsg("Canceled")
	}

	c.subConfig.Name = u.Message.Text // TODO validation
	c.update = c.updateSubName

	return c.textMsg("URL?")
}

func (c *Chat) updateSubName(u *telbot.Update) *telbot.MessageParams {
	if u.Message.Command() == "cancel" {
		c.update = c.updateIdle
		return c.textMsg("Canceled")
	}

	c.subConfig.URL = u.Message.Text // TODO validation
	c.update = c.updateSubURL

	return c.textMsg("Duration?")
}

func (c *Chat) updateSubURL(u *telbot.Update) *telbot.MessageParams {
	if u.Message.Command() == "cancel" {
		c.update = c.updateIdle
		return c.textMsg("Canceled")
	}

	duration, err := time.ParseDuration(u.Message.Text)
	if err != nil {
		return c.textMsg(fmt.Sprintf("could not parse duration: %s", err))
	}

	c.subConfig.Duration = duration
	name := c.subConfig.Name
	url := c.subConfig.URL

	c.sb.Subscribe(name, url, duration, c.handleAnomaly)

	c.update = c.updateIdle
	return c.textMsg("Created")
}

func (c *Chat) handleAnomaly(r *accessible.Result, err error) error {
	// TODO
	return nil
}

func (c *Chat) textMsg(t string) *telbot.MessageParams {
	return &telbot.MessageParams{
		ChatId: c.id,
		Text:   t,
	}
}

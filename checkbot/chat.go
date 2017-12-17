package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/cirias/accessible"
	"github.com/cirias/accessible/telbot"
)

type Chat struct {
	id     telbot.ChatId
	sb     *Subscriber
	params SubcribeParams
	update func(*telbot.Bot, *telbot.Update) *telbot.MessageParams
}

func NewChat(id telbot.ChatId, client *accessible.Client) *Chat {
	c := &Chat{
		id: id,
		sb: NewSubscriber(client),
	}
	c.update = c.updateIdle
	return c
}

func (c *Chat) Handle(bot *telbot.Bot, u *telbot.Update) {
	msg := c.update(bot, u)
	if msg == nil {
		return
	}

	if _, err := bot.SendMessage(msg); err != nil {
		log.Printf("could not send message: %v\n", err)
	}
}

func (c *Chat) updateIdle(_ *telbot.Bot, u *telbot.Update) *telbot.MessageParams {
	var msg *telbot.MessageParams
	cmd, arg := u.Message.Command()
	switch cmd {
	case "/sub":
		c.update = c.updateSubNew
		msg = c.textMsg("Name?")

	case "/ls":
		var buf bytes.Buffer
		c.sb.subs.Range(func(k, v interface{}) bool {
			s := v.(*Subscription)
			fmt.Fprintln(&buf, s)
			return true
		})
		msg = c.textMsg(fmt.Sprintf("subscriptions:\n%s", buf.String()))

	case "/log":
		sub, ok := c.sb.subs.Load(arg)
		if !ok {
			msg = c.textMsg("could not found")
			break
		}

		var buf bytes.Buffer
		sub.(*Subscription).History().Range(func(r *accessible.Result) bool {
			fmt.Fprintln(&buf, *r)
			return true
		})
		msg = c.textMsg(fmt.Sprintf("logs:\n%s", buf.String()))

	case "/rm":
		if err := c.sb.Unsubscribe(arg); err != nil {
			msg = c.textMsg(fmt.Sprintf("could not unsubscribe: %s", err))
			break
		}
		msg = c.textMsg("Removed")
	}
	return msg
}

func (c *Chat) updateSubNew(_ *telbot.Bot, u *telbot.Update) *telbot.MessageParams {
	if cmd, _ := u.Message.Command(); cmd == "cancel" {
		c.update = c.updateIdle
		return c.textMsg("Canceled")
	}

	c.params.Name = u.Message.Text // TODO validation
	c.update = c.updateSubName

	return c.textMsg("URL?")
}

func (c *Chat) updateSubName(_ *telbot.Bot, u *telbot.Update) *telbot.MessageParams {
	if cmd, _ := u.Message.Command(); cmd == "cancel" {
		c.update = c.updateIdle
		return c.textMsg("Canceled")
	}

	c.params.URL = u.Message.Text // TODO validation
	c.update = c.updateSubURL

	return c.textMsg("Duration?")
}

func (c *Chat) updateSubURL(bot *telbot.Bot, u *telbot.Update) *telbot.MessageParams {
	if cmd, _ := u.Message.Command(); cmd == "cancel" {
		c.update = c.updateIdle
		return c.textMsg("Canceled")
	}

	duration, err := time.ParseDuration(u.Message.Text)
	if err != nil {
		return c.textMsg(fmt.Sprintf("could not parse duration: %s", err))
	}

	c.params.Duration = duration

	h := func(r *accessible.Result, err error) error {
		m := c.textMsg("")
		if err != nil {
			m.Text = fmt.Sprintf("[ERROR] could not check: %s", err)
		} else {
			m.Text = fmt.Sprintf("[ERROR] failure check result: %v", r)
		}

		if _, err := bot.SendMessage(m); err != nil {
			log.Printf("could not send message: %v\n", err)
		}

		return nil
	}

	c.sb.Subscribe(c.params, h)

	c.update = c.updateIdle
	return c.textMsg("Created")
}

func (c *Chat) textMsg(t string) *telbot.MessageParams {
	return &telbot.MessageParams{
		ChatId: c.id,
		Text:   t,
	}
}

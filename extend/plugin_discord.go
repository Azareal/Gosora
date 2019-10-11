package extend

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	c "github.com/Azareal/Gosora/common"
)

func init() {
	c.Plugins.Add(&c.Plugin{UName: "discord", Name: "Discord", Author: "Azareal", Init: discordInit, Activate: discordActivate, Deactivate: discordDeactivate})
}

func discordValidate() error {
	webhook, ok := c.PluginConfig["DiscordWebhook"]
	if !ok || webhook == "" {
		return errors.New("You need to set a webhook to push to in config.json")
	}

	ev := c.PluginConfig["DiscordEvents"]
	if ev != "" && ev != "threads" && ev != "replies" {
		return errors.New("Invalid value for DiscordEvents. Can only be blank, 'threads' or 'replies'")
	}

	fidsRaw := c.PluginConfig["DiscordForums"]
	if fidsRaw != "" {
		for _, fidRaw := range strings.Split(fidsRaw, ",") {
			_, err := strconv.Atoi(fidRaw)
			if err != nil {
				return errors.New("Invalid integer found in DiscordForums")
			}
		}
	}

	return nil
}

func discordInit(plugin *c.Plugin) error {
	err := discordValidate()
	if err != nil {
		return err
	}
	plugin.AddHook("action_end_create_topic", discordEventTopic)
	plugin.AddHook("action_end_create_reply", discordEventReply)
	return nil
}

// A bit of validation to make sure the admin isn't forgetting something or telling Plugin Discord to do something absurd
func discordActivate(plugin *c.Plugin) error {
	return discordValidate()
}

func discordDeactivate(plugin *c.Plugin) {
	plugin.RemoveHook("action_end_create_topic", discordEventTopic)
	plugin.RemoveHook("action_end_create_reply", discordEventReply)
}

func discordEventTopic(args ...interface{}) (skip bool, rerr c.RouteError) {
	discordEvent(0, args[0].(int))
	return false, nil
}
func discordEventReply(args ...interface{}) (skip bool, rerr c.RouteError) {
	discordEvent(1, args[0].(int))
	return false, nil
}

type DiscordData struct {
	Username string         `json:"username"`
	Embeds   []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title  string             `json:"title"`
	Desc   string             `json:"description"`
	URL    string             `json:"url"`
	Author DiscordEmbedAuthor `json:"author"`
}

type DiscordEmbedAuthor struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Avatar string `json:"icon_url"`
}

func discordEvent(typ int, id int) {
	//fmt.Println("in discordEvent")
	ev := c.PluginConfig["DiscordEvents"]
	if (ev == "threads" && typ != 0) || (ev == "replies" && typ != 1) {
		return
	}

	var content, url string
	var topic *c.Topic
	var err error
	var createdBy int
	if typ == 0 {
		topic, err = c.Topics.Get(id)
		if err != nil {
			return
		}
		content = topic.Content
		createdBy = topic.CreatedBy
	} else {
		reply, err := c.Rstore.Get(id)
		if err != nil {
			return
		}
		content = reply.Content
		createdBy = reply.CreatedBy

		topic, err = reply.Topic()
		if err != nil {
			return
		}
	}
	url = topic.Link

	user, err := c.Users.Get(createdBy)
	if err != nil {
		return
	}

	fidsRaw := c.PluginConfig["DiscordForums"]
	if fidsRaw != "" {
		var hasForum = false
		for _, fidRaw := range strings.Split(fidsRaw, ",") {
			fid, err := strconv.Atoi(fidRaw)
			if err != nil {
				return
			}
			if fid == topic.ParentID {
				hasForum = true
			}
		}
		if !hasForum {
			return
		}
	}
	if len(content) > 100 {
		content = content[:97] + "..."
	}

	var client = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	var s string
	if c.Site.EnableSsl {
		s = "s"
	}
	var preURL = "http" + s + "://" + c.Site.URL

	var avatar = user.MicroAvatar
	if len(user.MicroAvatar) > 1 {
		if user.MicroAvatar[0] == '/' && user.MicroAvatar[1] != '/' {
			avatar = preURL + avatar
		}
	}

	author := DiscordEmbedAuthor{Name: user.Name, URL: preURL + user.Link, Avatar: avatar}
	embed := DiscordEmbed{Title: topic.Title, Desc: content, URL: preURL + url, Author: author}
	dat := DiscordData{Username: c.Site.Name, Embeds: []DiscordEmbed{embed}}
	data, err := json.Marshal(dat)
	if err != nil {
		c.LogWarning(err)
		return
	}

	//fmt.Println("before discord push")
	resp, err := client.Post(c.PluginConfig["DiscordWebhook"], "application/json", bytes.NewBuffer(data))
	var body string
	var respErr = func(err error) {
		log.Printf("Sent: %+v\n", string(data))
		log.Printf("Response: %+v\n", resp)
		if body != "" {
			log.Printf("Response Body: %+v\n", body)
		}
		c.LogWarning(err)
	}

	if err != nil {
		respErr(err)
		return
	}
	defer resp.Body.Close()

	// TODO: Cap the amount we read
	bBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		respErr(err)
		return
	}
	body = string(bBody)

	if resp.StatusCode != 200 {
		respErr(err)
		return
	}

	c.DebugLog("Pushed event to Discord")
	c.DebugLogf("Sent: %+v\n", string(data))
	c.DebugLogf("Response: %+v\n", resp)
	c.DebugLogf("Response Body: %+v\n", body)
}

// TODO: Add a settings page or something?

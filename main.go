package hipbot

import (
	"crypto/tls"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-xmpp"
)

const (
	groupchat = "groupchat"
)

type (
	// Config holds the configuration for the bot
	Config struct {
		JabberID string
		Nick     string
		FullName string
		Host     string
		Rooms    []string
		Password string
		Debug    bool
		TLS      *tls.Config
	}

	// Message represents an XMPP message
	Message struct {
		// Text of the message
		Text string
		// From contains the JID and Name of the sender separated by a /
		From string
	}

	// Filter provides a way to perform checks before invoking a Handler
	// Handlers will only be invoked if all Filters return true
	Filter func(m Message) (string, bool)

	handleMatcher struct {
		Filters []Filter
		Pattern string
		Handler Handler
	}

	// Bot represents the bot
	Bot struct {
		xmpp     *xmpp.Client
		config   Config
		stop     chan bool
		handlers handleMatchers
		help     Handler
		stopped  bool
		// Errors will be sent any non-nil error values encountered while listening for events
		Errors chan error
	}

	// Handler is passed a message and sends the text to HipChat if ok
	Handler func(m Message) string

	handleMatchers []handleMatcher
)

// Less sorts the longest strings first so we match more specifc -> less specific
func (h handleMatchers) Less(i, j int) bool {
	return len(h[i].Pattern) > len(h[j].Pattern)
}

// Len implements part of the Sort.Sort interface
func (h handleMatchers) Len() int {
	return len(h)
}

// Swap implements the sort.Sort interface
func (h handleMatchers) Swap(i, j int) {
	h[j], h[i] = h[i], h[j]
}

// heartbeat sends periodic pings to keep connection alive until the stop channel is closed
func (b *Bot) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		b.xmpp.PingC2S(b.config.JabberID, "")
		select {
		case <-ticker.C:
		case <-b.stop:
			ticker.Stop()
			return
		}
	}
}

// AddHelp registers a Handler that can be called if no other handlers match.
func (b *Bot) AddHelp(h Handler) {
	b.help = h
}

// AddHandler registers a Handler for callbacks that will be invoked when the pattern is matched.
// If Filters are passed, all must return true before a Handler is invoked
func (b *Bot) AddHandler(pattern string, h Handler, f ...Filter) {
	b.handlers = append(b.handlers, handleMatcher{Pattern: pattern, Handler: h, Filters: f})
	sort.Sort(b.handlers)
}

// SendRoom sends a message to a room.
func (b *Bot) SendRoom(msg string, room string) error {
	xmppMsg := xmpp.Chat{Text: msg, Type: groupchat, Remote: room}
	_, err := b.xmpp.Send(xmppMsg)
	return err
}

// SendUser sends a message privately to a user.
func (b *Bot) SendUser(msg string, user string) error {
	xmppMsg := xmpp.Chat{Text: msg, Type: "chat", Remote: user}
	_, err := b.xmpp.Send(xmppMsg)
	return err
}

func (b *Bot) send(msg xmpp.Chat) error {
	_, err := b.xmpp.Send(msg)
	return err
}

func from(remote string) string {
	components := strings.Split(remote, "/")
	if len(components) == 2 {
		return components[1]
	}
	return ""
}

// Sender returns the name of the message sender
func (m Message) Sender() string {
	return from(m.From)
}

// sentByMe checks if our name matches the sender
func (b *Bot) sentByMe(remote string) bool {
	return from(remote) == b.config.FullName
}

func (b *Bot) handle(xmppMsg xmpp.Chat) string {
	// handle case where the bot name is spelled with capital letter
	xmppMsg.Text = strings.TrimPrefix(xmppMsg.Text,strings.Title(b.config.Nick))
	msg := Message{
		Text: strings.TrimSpace(strings.TrimPrefix(xmppMsg.Text, b.config.Nick)),
		From: xmppMsg.Remote,
	}
	for _, h := range b.handlers {
		if strings.HasPrefix(msg.Text, h.Pattern) {
			msg.Text = strings.TrimSpace(strings.TrimPrefix(msg.Text, h.Pattern))
			for _, f := range h.Filters {
				if resp, ok := f(msg); !ok {
					return resp
				}
			}
			return h.Handler(msg)
		}
	}
	if b.help != nil {
		return b.help(msg)
	}
	return ""
}

func (b *Bot) toMe(msg xmpp.Chat) bool {
	if msg.Text == "" {
		return false
	}

	if b.sentByMe(msg.Remote) {
		return false
	}

	if msg.Type == "error" {
		return false
	}

	// only pass back group messages if they start with nick
	if msg.Type == groupchat {
		return strings.HasPrefix(strings.ToLower(msg.Text), strings.ToLower(b.config.Nick))
	}

	// direct message
	return true
}

// listen collects incoming XMPP events and performs callbacks
func (b *Bot) listen() {
	for {
		select {
		case <-b.stop:
			return
		default:
		}
		xmppMsg, err := b.xmpp.Recv()
		if b.stopped {
			return
		}
		if err != nil {
			select {
			// only send an error to the channel if someone is reading from it
			case b.Errors <- err:
			}
		}
		if b.config.Debug {
			fmt.Printf("DEBUG: %+v\n", xmppMsg)
		}
		if msg, ok := xmppMsg.(xmpp.Chat); ok {
			if b.toMe(msg) {
				resp := b.handle(msg)
				if resp != "" {
					msg.Text = resp
					b.send(msg)
				}
			}
		}
	}
}

// Start listens for XMPP messages and joins any rooms
func (b *Bot) Start() error {
	go b.listen()
	go b.heartbeat()
	for _, room := range b.config.Rooms {
		_, err := b.xmpp.JoinMUCNoHistory(room, b.config.FullName)
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop terminates the XMPP connection
func (b *Bot) Stop() {
	b.stopped = true
	close(b.stop)
	b.xmpp.Close()
	close(b.Errors)
}

// New generates a new bot
func New(cfg Config) (*Bot, error) {
	b := &Bot{
		config: cfg,
		stop:   make(chan bool),
		Errors: make(chan error),
	}
	options := &xmpp.Options{
		Host:      cfg.Host,
		User:      cfg.JabberID,
		Debug:     cfg.Debug,
		Password:  cfg.Password,
		Resource:  "bot",
		StartTLS:  false,
		NoTLS:     true,
		TLSConfig: cfg.TLS,
	}
	var err error
	b.xmpp, err = options.NewClient()
	return b, err
}

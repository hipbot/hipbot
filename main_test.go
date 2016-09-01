package hipbot

import (
	"testing"

	"github.com/mattn/go-xmpp"
)

const sorry = "Sorry, I don't understand your request"

func po(msg Message) string {
	return "po"
}

func pong(msg Message) string {
	return "pong"
}

func ping(msg Message) string {
	return "ping"
}

func pongping(msg Message) string {
	return "pongping"
}

func Help(msg Message) string {
	return sorry
}

func TestBotLogin(t *testing.T) {
	cfg, err := EnvConfig()
	if err != nil {
		t.Error(err)
		return
	}
	bot, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	bot.AddHandler("ping", pong)
	bot.AddHandler("pong", ping)
	bot.AddHandler("pi", po)
	bot.AddHandler("pingpong", pongping)
	err = bot.Start()
	if err != nil {
		t.Fatal(err)
	}
	HandlerTests := []struct {
		sent     string
		expected string
	}{
		{"ping", "pong"},
		{"pingpong", "pongping"},
		{"pong", "ping"},
		{"pi", "po"},
	}
	for _, test := range HandlerTests {
		msg := xmpp.Chat{Text: test.sent}
		resp := bot.handle(msg)
		if resp != test.expected {
			t.Errorf("did not receive expected response sent=%s recieved=%s expected=%s\n", test.sent, resp, test.expected)
			continue
		}
		t.Logf("test for callback with %s passed\n", test.sent)
	}
	msg := xmpp.Chat{Text: "invalid text"}
	resp := bot.handle(msg)
	if resp != "" {
		t.Errorf("received response to unmatched message before help added")
	} else {
		t.Log("expected empty result when sending unmatched message and help not added")
	}
	bot.AddHelp(Help)
	resp = bot.handle(msg)
	if resp != sorry {
		t.Errorf("help did not return expected text")
	} else {
		t.Logf("help worked correctly")
	}
	bot.Stop()
}

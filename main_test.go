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

func f1(msg Message) string {
	return "f1"
}

func f2(msg Message) string {
	return "f2"
}

func help(msg Message) string {
	return sorry
}

func passFilter(msg Message) (string, bool) {
	return "passed", true
}

func failFilter(msg Message) (string, bool) {
	return "filtered", false
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
	bot.AddHandler("f1", f1, passFilter)
	bot.AddHandler("f2", f2, failFilter)
	err = bot.Start()
	if err != nil {
		t.Fatal(err)
	}
	HandlerTests := []struct {
		sent        string
		expected    string
		description string
	}{
		{"ping", "pong", "medium string match"},
		{"pingpong", "pongping", "longer string match"},
		{"pong", "ping", "short string match"},
		{"pi", "po", "short string match"},
		{"f1", "f1", "filter used that passes"},
		{"f2", "filtered", "filter used that rejects"},
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
	bot.AddHelp(help)
	resp = bot.handle(msg)
	if resp != sorry {
		t.Errorf("help did not return expected text")
	} else {
		t.Logf("help worked correctly")
	}
	bot.Stop()
}

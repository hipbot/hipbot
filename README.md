hipbot
==
Implements a HipChat bot framework intended to make it easy to write simple bots.

The primary interface is to register one or more Handler funcs that have a signature of:

    func (m hipbot.Message) string

Handlers are registered using:

    AddHandler(pattern string, h Handler, f ...Filter)


When a message appears in a room that the bot is subscribed to, it will look for messages starting with the nickname of the bot.  If that is found, it will then look at the rest of the message and see if any patterns registered by `AddHandler` match.  If so, the Handler with the longest matching pattern will be called.  

The text passed to a Handler will have the bot nickname and the pattern it was registered with stripped out.  If any `Filter` functions were passed in during the `AddHandler` call, each of those will be called with the Message.  If any Filter returns false, the string from the Filter will be returned and the Handler will not be called.


If the returned string is not empty, hipbot will send the message.

Example usage:

    // pong returns the phrase "pong" every time
    func pong(m hipbot.Message) string {
        return "pong"
    }
    // uppercase returns the text back as uppercase
    func upper(m hipbot.Message) string {
      return strings.ToUpper(m.Text)
    }
    ...

    cfg := &hipbot.Config{ ... } // or use hipbot.EnvConfig
    bot, _ := hipbot.New(cfg)
    bot.Start() // join rooms, starts heartbeat and listens for messages
    bot.AddHandler("ping",pong)
    bot.AddHandler("upper",upper)
    ...
    bot.Stop() // will close XMPP connection and background goroutines


If the code above was ran with a bot nick of 'hal', interactions would look like:

    user> hal ping
    bot> pong

    user> hal upper froggy
    bot> FROGGY  

A help Handler can be registered with:

  bot.AddHelp(handler)

This would be executed when a user types a message starting with the bot nick that doesn't match any Handler patterns.

Example with a Filter for access control:

    // accessCheck validates the user is one of our allowed users
    func accessCheck(m hipbot.Message) (string, bool) {
      sender := m.Sender()
      for _, user := range allowedUsers {
        if user == sender {
          return "", true
        }
      }
      return "sorry, access denied", false
    }

    bot.AddHandler("ping",pong,accessCheck)

package main

import (
  "fmt"
  "github.com/joho/godotenv"
  "github.com/slack-go/slack"
  "github.com/slack-go/slack/slackevents"
  "github.com/slack-go/slack/socketmode"
  "log"
  "os"
  "strings"
)

func main() {
  err := godotenv.Load(".env.local")
  if err != nil {
    log.Fatal("Error loading .env.local file")
  }

  SLACK_APP_TOKEN := os.Getenv("SLACK_APP_TOKEN")
  SLACK_BOT_TOKEN := os.Getenv("SLACK_BOT_TOKEN")
  SLACK_CHANNNEL_TEST := os.Getenv("SLACK_CHANNNEL_TEST")

  api := slack.New(
    SLACK_BOT_TOKEN,
    slack.OptionAppLevelToken(SLACK_APP_TOKEN),
    slack.OptionDebug(true),
    slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile | log.LstdFlags)),
  )

  socketMode := socketmode.New(
    api,
    socketmode.OptionDebug(true),
    socketmode.OptionLog(log.New(os.Stdout, "socketMode: ", log.Lshortfile | log.LstdFlags)),
  )

  authTest, err := api.AuthTest()
  if err != nil {
    fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN is invalid: %v\n", err)
    os.Exit(1)
  }
  selfUserId := authTest.UserID

  go func() {
    for envelope := range socketMode.Events {
      switch envelope.Type {

      case socketmode.EventTypeEventsAPI:

        socketMode.Ack(*envelope.Request)

        eventPayload, _ := envelope.Data.(slackevents.EventsAPIEvent)
        switch eventPayload.Type {
        case slackevents.CallbackEvent:
          switch event := eventPayload.InnerEvent.Data.(type) {

          case *slackevents.MessageEvent:
            //https://api.slack.com/events/message
            if event.User != selfUserId && strings.Contains(event.Text, "hello") {
              _, _, err := api.PostMessage(
                event.Channel,
                slack.MsgOptionText(
                  fmt.Sprintf(":wave: hello <@%v> !", event.User),
                  false))
              if err != nil {
                log.Printf("Failed to reply: %v", err)
              }
            }

          case *slackevents.ReactionAddedEvent:
            //https://api.slack.com/events/reaction_added
            _, _, err := api.PostMessage(
              SLACK_CHANNNEL_TEST,
              slack.MsgOptionText(
                fmt.Sprintf(":%v: reaction added!", event.Reaction),
                false))
            if err != nil {
              log.Printf("Failed to reply: %v", err)
            }

          default:
            socketMode.Debugf("Skipped: %v", event)
          }

        default:
          socketMode.Debugf("unsupported Events API eventPayload received")
        }

      default:
        socketMode.Debugf("Skipped: %v", envelope.Type)
      }

    }
  }()

  socketMode.Run()
}
package xtelegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	tb "gopkg.in/tucnak/telebot.v2"
)

var telegramAlert *TelegramAlert

type TelegramAlert struct {
	ctx        context.Context
	ErrorLevel LogLevel
	ServerName string
	ID         int64
	Token      string
	tBot       *tb.Bot
}

func NewTelegramAlert(level LogLevel, serverName, id, token string) *TelegramAlert {
	telID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil
	}

	tBot, err := tb.NewBot(tb.Settings{
		Token:  token,
		URL:    tb.DefaultApiURL,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil
	}

	telegramAlert = &TelegramAlert{
		ctx:        context.Background(),
		ErrorLevel: level,
		ServerName: serverName,
		ID:         telID,
		Token:      token,
		tBot:       tBot,
	}
	return telegramAlert
}

func TelegramSendAlert(msg ...string) {
	if len(msg) == 1 {
		telegramAlert.AlertMsg(msg[0])
	} else {
		telegramAlert.AlertMsg(strings.Join(msg, ":"))
	}
}

func (t *TelegramAlert) AlertMsg(msg string) {
	if t == nil {
		return
	}

	timeLocal, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return
	}
	time.Local = timeLocal
	now := time.Now().Local().Format("2006-01-02 15:04:05")
	var (
		to      *tb.Chat
		message string
	)
	to = &tb.Chat{ID: t.ID}
	switch t.ErrorLevel {
	case INFO:
		message += fmt.Sprintln("\n 【LEVEL-INFO】,TS:  " + now)
	case WARN:
		message += fmt.Sprintln("\n 【LEVEL-WARN】,TS:  " + now)
	case CRASH:
		message += fmt.Sprintln("\n 【LEVEL-CRASH】,TS:  " + now)
	case ERROR:
		message += fmt.Sprintln("\n 【LEVEL-ERROR】,TS:  " + now)
	default:
		message += fmt.Sprintln("\n 【LEVEL-DEFAULT】,TS:  " + now)
	}
	message += fmt.Sprintf("\n【======Service: %s======】\n", t.ServerName) + msg
	g, ctx := errgroup.WithContext(t.ctx)
	g.Go(func() error {
		_, err := t.tBot.Send(to, message)
		if err != nil {
			return err
		}
		return err
	})
	err = g.Wait()
	if err != nil {
		ctx.Done()
	}
}

package xtelegram

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
	"time"
)

var (
	options = new(Options)
	logT    = new(LogT)
)

type Msg struct {
	l          LogLevel
	msg        string
	isAll      bool
	isTelegram bool
	isSentry   bool
}

type Options struct {
	Telegram *Telegram
	TSentry  *TSentry
}

type Telegram struct {
	ID   int64
	TBot *tb.Bot
}

type TSentry struct {
	Hub *sentry.Hub
}

type LogT struct {
	ServerName string
	Mode       string
	logs       logx.Logger // 日志输出
	msg        *Msg
}

func NewLogT(serverName, mode string, logger logx.Logger, myOptions func() *Options) *LogT {
	options = myOptions()
	logT = &LogT{
		logs:       logger,
		ServerName: serverName,
		Mode:       mode,
	}
	return logT
}

func Infof(msg string, arg ...any) *LogT {
	msgStu := &Msg{
		isAll: true,
		msg:   fmt.Sprintf(msg, arg...),
		l:     INFO,
	}
	logT.logs.Infof(msgStu.msg)
	logT.msg = msgStu
	return logT
}

func Info(msg string) *LogT {
	msgStu := &Msg{
		isAll: true,
		msg:   msg,
		l:     INFO,
	}
	logT.logs.Infof(msgStu.msg)
	logT.msg = msgStu
	return logT
}

func Errorf(msg string, arg ...any) *LogT {
	msgStu := &Msg{
		isAll: true,
		msg:   fmt.Sprintf(msg, arg...),
		l:     ERROR,
	}
	logT.logs.Errorf(msgStu.msg)
	logT.msg = msgStu
	return logT
}

func Error(msg string) *LogT {
	msgStu := &Msg{
		isAll: true,
		msg:   msg,
		l:     ERROR,
	}
	logT.logs.Errorf(msgStu.msg)
	logT.msg = msgStu
	return logT
}

func Debugf(msg string, arg ...any) *LogT {
	msgStu := &Msg{
		isAll: true,
		msg:   fmt.Sprintf(msg, arg...),
		l:     DEBUG,
	}
	logT.logs.Debugf(msgStu.msg)
	logT.msg = msgStu
	return logT
}

func (t *LogT) Telegram() *LogT {
	t.msg.isAll = false
	if options.Telegram == nil || options.Telegram.TBot == nil {
		t.msg.isTelegram = false
		return t
	} else {
		t.msg.isTelegram = true
		return t
	}
}

func (t *LogT) Sentry() *LogT {
	t.msg.isAll = false
	if options.TSentry == nil || options.TSentry.Hub == nil {
		t.msg.isSentry = false
	} else {
		t.msg.isSentry = true
	}
	return t
}

func (t *LogT) Send() *LogT {
	if t.msg == nil || options == nil {
		return t
	}
	if t.msg.isAll || t.msg.isTelegram {
		if options.Telegram != nil && options.Telegram.TBot != nil {
			// 复制 t.msg.msg，防止多次 Send 调用时被覆盖
			currentMsg := t.msg.msg
			go func() {
				timeLocal, err := time.LoadLocation("Asia/Shanghai")
				if err != nil {
					return
				}
				now := time.Now().In(timeLocal).Format("2006-01-02 15:04:05")
				var message string
				switch t.msg.l {
				case INFO:
					message = fmt.Sprintf("【INFO】")
				case DEBUG:
					message = fmt.Sprintf("【DEBUG】")
				case ERROR:
					message = fmt.Sprintf("【ERROR】")
				default:
					message = fmt.Sprintf("【DEFAULT】")
				}
				message = strings.Join([]string{
					message,
					fmt.Sprintf("【%s】", t.Mode),
					fmt.Sprintf("TS: %s ", now),
					fmt.Sprintf("ServerName: %s ", t.ServerName),
					fmt.Sprintf(""),
					currentMsg, // 使用复制的日志信息
					"",
				}, "\n\t")
				g, ctx := errgroup.WithContext(context.Background())
				g.Go(func() error {
					_, err := options.Telegram.TBot.Send(&tb.Chat{ID: options.Telegram.ID}, message)
					if err != nil {
						return err
					}
					return err
				})

				err = g.Wait()
				if err != nil {
					ctx.Done()
				}
			}()
		}
	}
	if t.msg.isAll || t.msg.isSentry {
		if options.TSentry != nil {
			go func() {
				options.TSentry.Hub.CaptureMessage(t.msg.msg)
				sentry.Flush(2 * time.Second)
			}()
		}
	}
	return logT
}

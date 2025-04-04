package security

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v4"
	llm "s-belichenko/ilovaiskaya2-bot/cmd/llm"
	hndls "s-belichenko/ilovaiskaya2-bot/internal/handlers"
	intLog "s-belichenko/ilovaiskaya2-bot/internal/logger"
)

func AllPrivateChatsMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Chat().Type != tele.ChatPrivate && c.Chat().Type != tele.ChatChannelPrivate {
			log.Warn(
				fmt.Sprintf("Попытка использовать %q в чате типа %q", getCommandName(c.Message()), c.Chat().Type), intLog.LogContext{"message": c.Message()})
			if TeleID(c.Chat().ID) == config.HouseChatId {
				err := c.Reply(fmt.Sprintf("Используйте команду %q в личной переписке с ботом.", getCommandName(c.Message())))
				if err != nil {
					log.Error(
						fmt.Sprintf("Не удалось посоветовать использовать личную переписку с ботом: %v", err),
						intLog.LogContext{"message": c.Message()})
				}
			}
			return nil
		}

		return next(c)
	}
}

func HomeChatMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Chat().Type != tele.ChatGroup && c.Chat().Type != tele.ChatSuperGroup {
			log.Warn(fmt.Sprintf(
				"Попытка использовать %q в чате типа %q", getCommandName(c.Message()), c.Chat().Type,
			), intLog.LogContext{"message": c.Message()})
			return nil
		}

		if TeleID(c.Chat().ID) != config.HouseChatId {
			log.Warn(fmt.Sprintf(
				"Попытка использовать %q вне домового чата, чат: %d", getCommandName(c.Message()), c.Chat().ID,
			), intLog.LogContext{
				"message": c.Message(),
			})
			return nil
		}

		return next(c)
	}
}

func AdminChatMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Chat().Type != tele.ChatGroup && c.Chat().Type != tele.ChatSuperGroup {
			log.Warn(fmt.Sprintf(
				"Попытка использовать команду %q в чате типа %q", getCommandName(c.Message()), c.Chat().Type,
			), intLog.LogContext{
				"message": c.Message(),
			})
			return nil
		}

		if TeleID(c.Chat().ID) != config.AdministrationChatID {
			log.Warn(fmt.Sprintf(
				"Попытка использовать команду %q в чате %d", getCommandName(c.Message()), c.Chat().ID),
				intLog.LogContext{"message": c.Message()})
			return nil
		}

		if member, err := c.Bot().ChatMemberOf(c.Chat(), c.Sender()); err != nil {
			log.Error(
				fmt.Sprintf("Не удалось получить информацию об отправителе %q команды %q", hndls.GetGreetingName(c.Sender()), getCommandName(c.Message())),
				intLog.LogContext{"user_id": c.Sender().ID})
			return nil
		} else {
			if (tele.Creator != member.Role) && (tele.Administrator != member.Role) {
				link := fmt.Sprintf("<a href=%q>ссылка</a>", hndls.GenerateMessageLink(c.Chat(), c.Message().ID))
				reportMessage := fmt.Sprintf(
					"Хакир детектед! Пользователь %q попытался использовать команду %q, ссылка: %s",
					hndls.GetGreetingName(c.Sender()), getCommandName(c.Message()), link,
				)
				adminChat := &tele.Chat{ID: int64(config.AdministrationChatID)}
				_, _ = c.Bot().Send(adminChat, reportMessage, tele.ModeHTML)
			}
		}

		return next(c)
	}
}

func KeysCommandMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if !isBotHouse(c) {
			cantSpeakPhrase := llm.GetCantSpeakPhrase()
			if "" != cantSpeakPhrase {
				if !strings.HasSuffix(cantSpeakPhrase, ".") &&
					!strings.HasSuffix(cantSpeakPhrase, "!") &&
					!strings.HasSuffix(cantSpeakPhrase, "?") {
					cantSpeakPhrase += "."
				}
				// TODO: Через очереди записывать команды не в тех местах и удалять их по истечении некоего времени.
				//  Писать также куда-то злоупотребляющих командой не в тех местах? Писать вообще все команды куда-либо?
				//  Использовать DeleteAfter()?
				err := c.Reply(fmt.Sprintf(
					"%s @%s, попробуйте использовать команду в теме \"Оффтоп.\"",
					cantSpeakPhrase, c.Sender().Username,
				))
				if err != nil {
					log.Error(fmt.Sprintf("Бот не смог рассказать об ограничениях команды /keys: %v", err), nil)
				}
			}
			return nil
		}

		return next(c)
	}
}

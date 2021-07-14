package main

import (
	"fmt"
	"github.com/shodikhuja83/news/config"
	api_parser "github.com/shodikhuja83/news/internal/api-parser"
	"github.com/shodikhuja83/news/internal/db"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/mailru/dbr"
)

var (
	dbConn *dbr.Connection
	cfg *config.Config
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		ForceQuote:      true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(log.TraceLevel)
	cfg = config.InitConfig()

	var err error
	dbConn, err = db.CreatePostgresClient(cfg)
	if err != nil {
		log.Fatalln(err)
	}
}

var (
	helpMessage = "Снова приветули! Это <b>Newsman</b>, я расскажу тебе немного о том, как я работаю!\n\n" +
		"Я запоминаю все интересные темы, которые ты выбрал, общаясь со мной в чатике. " +
		"Раз в час я провожу сканирование всех интересных новостей на диком западе. " +
		"Естественно сканирую новости не я саморучно, а команда индусов-операционистов!\n\n" +
		"Эх, я кажется немного увлекся. Кстати, я не работаю в тех странах где запрещена работорговля, <b>даже не спрашивай почему!</b>\n\n" +
		"Давай ка я напомню тебе как пользоваться мной:\n" +
		"<i>/subscribe</i> - позволяет вывести список доступных тем и произвести подписку\n" +
		"<i>/unsubscribe</i> - позволяет отменить уже созданную подписку на какую либо тему\n" +
		"<i>/help</i> - прочитать мой прекрасный монолог ещё раз\n\n"
	startMessage = "Добро пожаловать на страницу самого лучшего новостного бота <b>Newsman!</b>\n\nТебе доступны две команды из списка:\n" +
	"<i>/subscribe</i> - позволяет вывести список доступных тем и произвести подписку\n" +
	"<i>/unsubscribe</i> - позволяет отменить уже созданную подписку на какую либо тему\n\n" +
	"Если снова понадобится помощь - просто набери <i>/help</i>!"
	subscribeMessage = "Соскучился по свежим новостям, <b>%s</b>? Мы очень рады что ты пользуешься нашим ботом ;)\n\n" +
		"\tВыбери одну из указанных тем для подписки:"
	confirmSubscribeMessage = "%s? Хороший выбор, моя любимая тема, часто обсуждаем с ребятами в качалке ^-^\nМы уже совсем скоро закидаем тебя новостями по этой теме!"
	unsubscribeMessageNoSubscribes = "Дружище, а тебе даже не от чего отписываться!!"
	unsubscribeMessage = "Воу, полегче, ковбой. Уверен что хочешь отказаться от таких классных тем?\n%s\n" +
		"Просто <b>нажми на кнопку</b> с темой от которой хочет отписаться и мы всё сделаем!\n" +
		"Помни что ты всегда сможешь снова запросить у нас новости по этой теме!"
	unsubscribeConfirmMessage = "Мы успешно отписали тебя от <b>%s</b>. Помни, что ты всегда можешь возобновить подписку на данную тему! ;)"
	newArticleMessage = "<b>Раздел:</b> %s\n<b>Название:</b> %s\n\n%s\n" +
		"<b>От:</b> %s\n" +
		"<i>%s</i>"
)

var subscribeKeyboardMarkup = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Бизнес", "business"),
		tgbotapi.NewInlineKeyboardButtonData("Развлечение", "entertainment"),
		tgbotapi.NewInlineKeyboardButtonData("Общее", "general"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Здоровье", "health"),
		tgbotapi.NewInlineKeyboardButtonData("Наука", "science"),
		tgbotapi.NewInlineKeyboardButtonData("Спорт", "sports"),
		tgbotapi.NewInlineKeyboardButtonData("Технологии", "technology"),
	),
)

func formKeyboardMarkupForUnsubscribe(tags []string) tgbotapi.InlineKeyboardMarkup {
	var keyboardButtons []tgbotapi.InlineKeyboardButton
	for _, tag := range tags {
		keyboardButtons = append(keyboardButtons, tgbotapi.NewInlineKeyboardButtonData(tag, "unsub " + tag))
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(keyboardButtons...))
}

func mapModelsToRequestMap(models []db.SubscribeModel) (requestMap map[string][]int64, tagKeys []string) {
	requestMap = make(map[string][]int64)
	tagKeys = make([]string, 0)
	for _, model := range models {
		requestMap[model.Tag] = append(requestMap[model.Tag], model.UserId)
	}
	for key, _ := range requestMap {
		log.Infof("adding %s key to tagKeys", key)
		tagKeys = append(tagKeys, key)
	}
	return
}

func main() {
	log.Println(cfg.TgBotApiToken)
	bot, err := tgbotapi.NewBotAPI(cfg.TgBotApiToken)
	if err != nil {
		log.Fatalln("bot creation error", err)
	}
	log.Infof("%s bot was initialized", bot.Self.FirstName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(cfg.TgWebhookUrl))
	if err != nil {
		log.Fatalln("webhook error", err)
	}

	parserClient := api_parser.CreateParserClient(&http.Client{})

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":8080", nil)
	log.Println("Starting to listen and serve :8080")

	var s = dbConn.NewSession(nil)

	go func() {
		for _ = range time.Tick(time.Second * 45) {
			log.Infoln("Starting ticker work ;O")

			subscribeModels, err := db.SelectAllSubscribes(s)
			if err != nil {
				log.Errorln(err)
			}
			log.Infof("Got %d subscribes", len(subscribeModels))

			requestMap, tagsFromRequestMap := mapModelsToRequestMap(subscribeModels)
			log.Infof("requestMap:\n%+v\ntagsFromRequestMap[%d]:\n%+v", requestMap, len(tagsFromRequestMap),tagsFromRequestMap)

			responseMap := parserClient.GetArticlesByTags(tagsFromRequestMap)

			for tag, response := range responseMap {
				log.Printf("parsing response for tag %s, found %d articles\n", tag, len(response.Articles))
				if len(response.Articles) == 0 {
					log.Warnln("Skipping sending article cause of 0 response")
					continue
				}
				log.Infof("Forming message for users: %+v", requestMap[tag])
				for _, userId := range requestMap[tag] {
					for _, article := range response.Articles {
						contains, err := db.IsSubscribeContainsArticle(s, userId, tag, article.Title)
						if err != nil {
							log.Errorln("db err IsSubscribeContainsArticle:", err)
							continue
						}
						if !contains {
							err = db.AddReadenArticleToSubscribe(s, userId, tag, article.Title)
							if err != nil {
								log.Errorln("db err AddReadenArticle:", err)
								continue
							}
							log.Infof("Forming message for %d", userId)
							msgToSend := tgbotapi.NewMessage(userId, fmt.Sprintf(newArticleMessage, tag,
								article.Title, article.Description, article.Author, article.Url))
							msgToSend.ParseMode = "html"
							_, err = bot.Send(msgToSend)
							if err != nil {
								log.Errorln("error sending message ticker:", err)
								continue
							}
							break
						}
					}
				}
			}
		}
	}()

	for update := range updates {
		if update.CallbackQuery != nil {
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID,update.CallbackQuery.Data))

			var callbackMsg tgbotapi.MessageConfig
			if strings.HasPrefix(update.CallbackQuery.Data, "unsub ") {
				log.Infof("Trying to delete subscribe for tag %s, data %s", strings.TrimPrefix(update.CallbackQuery.Data, "unsub "), update.CallbackQuery.Data)
				if err = db.DeleteSubscribeFromSubscribes(s, update.CallbackQuery.Message.Chat.ID, strings.TrimPrefix(update.CallbackQuery.Data, "unsub ")); err != nil {
					log.Errorln("error sending message:", err)
					continue
				}
				callbackMsg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(unsubscribeConfirmMessage, strings.TrimPrefix(update.CallbackQuery.Data, "unsub ")))
				callbackMsg.ParseMode = "html"
				_, err = bot.Send(callbackMsg)
				if err != nil {
					log.Errorln("error sending message:", err)
					continue
				}
				continue
			}

			tagToSubscribe := update.CallbackQuery.Data

			err = db.InsertNewSubscribeToSubscribes(s, db.SubscribeModel{
				UserId:         int64(update.CallbackQuery.From.ID),
				Tag:            tagToSubscribe,
			})
			if err != nil {
				log.Errorln("Unexpected error from DB:", err)
				continue
			}

			callbackMsg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(confirmSubscribeMessage, tagToSubscribe))
			_, err = bot.Send(callbackMsg)
			if err != nil {
				log.Errorln("error sending message:", err)
				continue
			}
		}
		if update.Message == nil {
			continue
		}

		log.Printf("Got a new message from %d:%s\n[%s]", update.Message.Chat.ID, update.Message.Chat.FirstName, update.Message.Text)

		var msgToSend tgbotapi.MessageConfig
		switch update.Message.Text {
		case "/start":
			msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, startMessage)
			msgToSend.ParseMode = "html"
		case "/subscribe":
			msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(subscribeMessage, update.Message.Chat.FirstName))
			msgToSend.ParseMode = "html"
			msgToSend.ReplyMarkup = subscribeKeyboardMarkup
		case "/unsubscribe":
			tags, err := db.SelectUsersSubscribeTags(s, update.Message.Chat.ID)
			if err != nil {
				log.Errorln("db error:", err)
				continue
			}
			if len(tags) > 0 {
				msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(unsubscribeMessage, strings.Join(tags, `, `)))
				msgToSend.ParseMode = "html"
				msgToSend.ReplyMarkup = formKeyboardMarkupForUnsubscribe(tags)
			} else {
				msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, unsubscribeMessageNoSubscribes)
			}
		case "/help":
			msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
			msgToSend.ParseMode = "html"
		default:
			msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, "Что то на богатом")
		}

		_, err = bot.Send(msgToSend)
		if err != nil {
			log.Errorln("error sending message:", err)
			continue
		}
	}

}

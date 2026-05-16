package service

import (
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

func BuildReport(report model.Report) *model.Message {
	sb := strings.Builder{}
	sb.WriteString("Репорт по ссылкам, по которым не удалось получить обновления\n")
	for _, linkError := range report.LinkUpdateErrors {
		sb.WriteString("Ссылка: " + linkError.Link + "\n")
		sb.WriteString("Ошибка: " + linkError.Err + "\n")
		sb.WriteString("Сообщение: " + linkError.Message + "\n")
	}
	return model.NewMessage(sb.String())
}

func BuildUpdate(update model.LinkUpdate) *model.Message {
	return model.NewMessage("Обновление по ссылке: " + update.Link + "\n" +
		update.Update.Title + "\n" +
		"Автор: " + update.Update.Author + "\n" +
		"Время создания: " + update.Update.TimeCreated.String() + "\n" +
		"Превью: " + update.Update.Description + "\n")
}

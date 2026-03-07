package mapper

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/scrapper"
)

func ToDomainLinkUpdate(update *bot.LinkUpdate) *model.LinkUpdate {
	return model.NewLinkUpdate(update.Id, update.Link, update.Description, update.TgChatIDS)
}

func ToProtoLinkUpdate(update *model.LinkUpdate) *bot.LinkUpdate {
	return &bot.LinkUpdate{
		Id:          update.Id,
		Link:        update.Url,
		Description: update.Description,
		TgChatIDS:   update.TgChatIds,
	}
}

func ToProtoLinkResponse(link *model.Link) *scrapper.LinkResponse {
	return &scrapper.LinkResponse{
		Id:   link.Id,
		Url:  link.Link,
		Tags: link.Tags,
	}
}

func ToDomainLinkResponse(link *scrapper.LinkResponse) *dto.LinkResponse {
	return &dto.LinkResponse{
		Id:   link.Id,
		Link: link.Url,
		Tags: link.Tags,
	}
}
func ToDomainAddLinkRequest(req *scrapper.AddLinkRequest) dto.AddLinkRequest {
	return dto.AddLinkRequest{
		Link: req.Link,
		Tags: req.Tags,
	}
}

func ToDomainRemoveLinkRequest(req *scrapper.RemoveLinkRequest) dto.DeleteLinkRequest {
	return dto.DeleteLinkRequest{Link: req.Link}
}

func ToProtoChatID(chatID int64) *scrapper.ChatID {
	return &scrapper.ChatID{Id: chatID}
}

func ToProtoAddLinkRequest(req *dto.AddLinkRequest) *scrapper.AddLinkRequest {
	return &scrapper.AddLinkRequest{
		Link: req.Link,
		Tags: req.Tags,
	}
}

func ToProtoDeleteLinkRequest(req *dto.DeleteLinkRequest) *scrapper.RemoveLinkRequest {
	return &scrapper.RemoveLinkRequest{
		Link: req.Link,
	}
}

package mapper

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/scrapper"
)

func ToDomainLinkUpdate(update *bot.LinkUpdate) *domain.LinkUpdate {
	return domain.NewLinkUpdate(update.Id, update.Link, update.Description, update.TgChatIDS)
}

func ToProtoLinkUpdate(update *domain.LinkUpdate) *bot.LinkUpdate {
	return &bot.LinkUpdate{
		Id:          update.Id,
		Link:        update.Url,
		Description: update.Description,
		TgChatIDS:   update.TgChatIds,
	}
}

func ToProtoLinkResponse(link *domain.Link) *scrapper.LinkResponse {
	return &scrapper.LinkResponse{
		Id:   link.Id,
		Url:  link.Link,
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

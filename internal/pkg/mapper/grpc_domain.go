package mapper

import (
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/scrapper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toDomainUpdate(update *bot.Update) model.Update {
	return model.Update{
		Title:        update.Title,
		Author:       update.Author,
		TimeCreated:  update.TimeCreated.AsTime(),
		LastModified: update.LastModified.AsTime(),
		Description:  update.Description,
	}
}

func toProtoUpdate(update model.Update) *bot.Update {
	return &bot.Update{
		Title:        update.Title,
		Author:       update.Author,
		TimeCreated:  timestamppb.New(update.TimeCreated),
		LastModified: timestamppb.New(update.LastModified),
		Description:  update.Description,
	}
}

func ToDomainLinkUpdate(update *bot.LinkUpdate) *model.LinkUpdate {
	return model.NewLinkUpdate(update.Id, update.Link, toDomainUpdate(update.Update), update.TgChatIDs)
}

func ToProtoLinkUpdate(update *model.LinkUpdate) *bot.LinkUpdate {
	return &bot.LinkUpdate{
		Id:        update.ID,
		Link:      update.Link,
		Update:    toProtoUpdate(update.Update),
		TgChatIDs: update.TgChatIDs,
	}
}

func ToProtoLinkResponse(link *model.Link) *scrapper.LinkResponse {
	return &scrapper.LinkResponse{
		Id:   link.ID,
		Link: link.Link,
		Tags: link.Tags,
	}
}

func ToDomainLinkResponse(link *scrapper.LinkResponse) *dto.LinkResponse {
	return &dto.LinkResponse{
		ID:   strconv.FormatInt(link.Id, 10),
		Link: link.Link,
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

func ToProtoGetLinksReq(chatID int64, tag string) *scrapper.GetLinksReq {
	return &scrapper.GetLinksReq{Id: chatID, Tag: tag}
}

func ToProtoAddLinkRequest(chatID int64, req *dto.AddLinkRequest) *scrapper.AddLinkRequest {
	return &scrapper.AddLinkRequest{
		ChatID: chatID,
		Link:   req.Link,
		Tags:   req.Tags,
	}
}

func ToProtoDeleteLinkRequest(chatID int64, link string) *scrapper.RemoveLinkRequest {
	return &scrapper.RemoveLinkRequest{
		ChatID: chatID,
		Link:   link,
	}
}

func ToDomainError(linkUpdateErr *bot.LinkUpdateError) *model.LinkUpdateError {
	return &model.LinkUpdateError{
		Link:    linkUpdateErr.Link,
		Err:     linkUpdateErr.Err,
		Message: linkUpdateErr.Message,
		LinkID:  linkUpdateErr.LinkId,
	}
}

func ToProtoError(linkUpdateErr *model.LinkUpdateError) *bot.LinkUpdateError {
	return &bot.LinkUpdateError{
		Link:    linkUpdateErr.Link,
		Err:     linkUpdateErr.Err,
		Message: linkUpdateErr.Message,
		LinkId:  linkUpdateErr.LinkID,
	}
}

func ToDomainReport(report *bot.Report) model.Report {
	var errors []*model.LinkUpdateError
	for _, err := range report.Errors {
		errors = append(errors, ToDomainError(err))
	}

	return model.Report{
		ID:               report.Id,
		ChatID:           report.ChatId,
		LinkUpdateErrors: errors,
	}
}

func ToProtoReport(report *model.Report) *bot.Report {
	var errors []*bot.LinkUpdateError
	for _, err := range report.LinkUpdateErrors {
		errors = append(errors, ToProtoError(err))
	}

	return &bot.Report{
		Id:     report.ID,
		ChatId: report.ChatID,
		Errors: errors,
	}
}

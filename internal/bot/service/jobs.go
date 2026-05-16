package service

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"

type UpdateJob struct {
	model.LinkUpdate
	Result chan error
}

type ReportJob struct {
	model.Report
	Result chan error
}

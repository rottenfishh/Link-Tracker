package out

import pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/scrapper"

type ScrapperGrpcClient struct {
	pb.ScrapperServiceClient
}

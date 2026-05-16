package valkey_test

import (
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

//nolint:unparam // thats how mocks work man
func expectGetLinksByChatID(mock sqlmock.Sqlmock, chatID int64, linkURL string) {
	rows := sqlmock.NewRows([]string{"ID", "link", "domain", "last_updated"}).
		AddRow(1, linkURL, "github", time.Now())
	mock.ExpectQuery("SELECT").
		WithArgs(chatID).
		WillReturnRows(rows)
}

func expectAddLink(mock sqlmock.Sqlmock, chatID int64, linkURL string) {
	mock.ExpectBegin()
	// INSERT INTO links ... RETURNING ...
	mock.ExpectQuery("INSERT INTO links").
		WillReturnRows(sqlmock.NewRows([]string{
			"ID", "link", "domain", "last_updated", "formatted_link", "title",
		}).AddRow(1, linkURL, "github", time.Now(), linkURL, "Mock Title"))
	// INSERT INTO chat_link
	mock.ExpectExec("INSERT INTO chat_link").
		WithArgs(chatID, int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

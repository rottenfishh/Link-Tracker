package database

import "embed"

//go:embed scrapper/*.sql
var ScrapperMigrations embed.FS

//go:embed bot/*sql
var BotMigrations embed.FS

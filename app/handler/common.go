package handler

import (
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/view"
	"github.com/pmmp/CrashArchive/app/webhook"
)

type Common struct {
	DB      *database.DB
	WH      *webhook.Webhook
	View    view.Views
	Assets  string
	Reports string
}

package mysql

import (
	"context"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var (
	BoilCtx      = boil.WithDebug(context.Background(), true)
	BoilCtxNoLog = boil.WithDebug(context.Background(), false)
)

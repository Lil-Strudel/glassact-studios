package data

import (
	"time"

	"github.com/go-jet/jet/v2/postgres"
)

type StandardTable struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

func Citext(val string) postgres.StringExpression {
	return postgres.StringExp(postgres.Raw("$1::citext", map[string]any{
		"$1": val,
	}))
}

func STPoint(longitude, latitude float64) postgres.StringExpression {
	return postgres.StringExp(postgres.Raw(
		"ST_SetSRID(ST_MakePoint($1, $2), 4326)::GEOGRAPHY",
		map[string]any{
			"$1": longitude,
			"$2": latitude,
		},
	))
}

func STLongitude(locationExpr postgres.Expression) postgres.Expression {
	return postgres.Func("ST_X", postgres.CAST(locationExpr).AS("GEOMETRY"))
}

func STLatitude(locationExpr postgres.Expression) postgres.Expression {
	return postgres.Func("ST_Y", postgres.CAST(locationExpr).AS("GEOMETRY"))
}

func Now() postgres.Expression {
	return postgres.Raw("now()")
}

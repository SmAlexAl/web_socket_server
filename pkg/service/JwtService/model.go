package JwtService

import "database/sql"

type ProfileDto struct {
	Dci         string         `db:"dci"`
	DeveciId    string         `db:"deviceId"`
	InstallId   string         `db:"installId"`
	Udid        sql.NullString `db:"udid"`
	AccountId   sql.NullString `db:"accountId"`
	ProfileId   sql.NullString `db:"profileId"`
	ProfileName sql.NullString `db:"profileName"`
	ClientId    sql.NullString `db:"clientId"`
	ClanId      sql.NullString `db:"clanId"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type TokenData struct {
	ApplicationPackageName string
	InstallId              string
	Udid                   string
	Platform               string
}

//dci.id,
//dci.deviceId,
//dci.installId,
//d.udid,
//a.id as accountId,
//p.id as profileId,
//c.id as clientId,

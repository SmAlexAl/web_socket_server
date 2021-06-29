package JwtService

import (
	"database/sql"
	"log"
)

func GetProfileData(conn *sql.DB, tokenData TokenData) *ProfileDto {
	var profileDto *ProfileDto

	queryPattern := `
            SELECT
                dci.id,
                dci.deviceId,
                dci.installId,
                d.udid,
                a.id as accountId,
                p.id as profileId,
                c.id as clientId
            FROM DeviceClientInstall dci
            INNER JOIN ClientVersion cv ON dci.clientVersionId = cv.id
            INNER JOIN Device d ON dci.deviceId = d.id AND d.udid = ?
            INNER JOIN Client c ON cv.clientId = c.id
            LEFT JOIN Account a on dci.accountId = a.id
            INNER JOIN Profile p on p.accountId = a.id
            WHERE 1
                AND dci.packageName = ?
                AND dci.installId = ?
                AND d.udid = ?
                AND d.platform = ?				
    `

	rows, err := conn.Query(
		queryPattern,
		tokenData.Udid,
		tokenData.ApplicationPackageName,
		tokenData.InstallId,
		tokenData.Udid,
		tokenData.Platform,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err = rows.Close()

		if err != nil {
			log.Fatal(err)
		}
	}()
	for rows.Next() {
		profileDto := new(ProfileDto)
		err := rows.Scan(
			&profileDto.Dci,
			&profileDto.DeveciId,
			&profileDto.InstallId,
			&profileDto.Udid,
			&profileDto.AccountId,
			&profileDto.ProfileId,
			&profileDto.ClientId,
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return profileDto
}

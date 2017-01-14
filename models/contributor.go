package models

import (
	"errors"
	"ims-release/database"
	"time"
)

type Contributor struct {
	Id        uint32    `json:"id"`
	Name      string    `json:"name"`
	Biography string    `json:"biography"`
	CreatedAt time.Time `json:"createdAt"`
}

const (
	t_contributors string = "`contributors`"
	Cc_id          string = "`id`"
	Cc_name        string = "`name`"
	Cc_biography   string = "`biography`"
	Cc_created_at  string = "`created_at`"

	Cmax_len_name      = 65535
	Cmax_len_biography = 65535
)

var (
	ErrNoSuchContributor = errors.New("Could not find contributor.")
)

func NewContributor(name, biography string, tm time.Time) Contributor {
	return Contributor{
		0,
		name,
		biography,
		tm,
	}
}

func FindContributor(db database.DB, id uint32) (Contributor, error) {
	c := Contributor{}
	const query = "SELECT " + Cc_name + ", " + Cc_biography + ", " +
		Cc_created_at + " FROM " + t_contributors + " WHERE " + Cc_id + " = ?"

	row := db.QueryRow(query, id)
	err := row.Scan(&c.Name, &c.Biography, &c.CreatedAt)
	if err == database.ErrNoRows {
		return Contributor{}, ErrNoSuchContributor
	} else if err != nil {
		return Contributor{}, err
	}
	c.Id = id
	return c, nil
}

func ListContributors(db database.DB) ([]Contributor, error) {
	contributors := []Contributor{}

	const query = "SELECT " + Cc_id + ", " + Cc_name + ", " +
		Cc_biography + ", " + Cc_created_at + " FROM " + t_contributors

	rows, err := db.Query(query)
	if err != nil {
		return []Contributor{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var c Contributor
		err = rows.Scan(&c.Id, &c.Name, &c.Biography, &c.CreatedAt)
		if err != nil {
			return contributors, err
		}

		contributors = append(contributors, c)
	}
	err = rows.Err()
	return contributors, err
}

func (p *Contributor) Validate() error {
	if len(p.Name) > Cmax_len_name || len(p.Biography) > Cmax_len_biography {
		return ErrFieldTooLong
	}

	return nil
}

func SaveContributor(db database.DB, c Contributor) (Contributor, error) {
	validErr := c.Validate()
	if validErr != nil {
		return c, validErr
	}

	const query = "INSERT INTO " + t_contributors + " (" +
		Cc_name + ", " + Cc_biography + ", " + Cc_created_at + ") VALUES (?, ?, ?, ?, ?)"

	res, err := db.Exec(query, c.Name, c.Biography, c.CreatedAt)
	if err != nil {
		return c, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return c, err
	}
	c.Id = uint32(id)
	return c, nil
}

func UpdateContributor(db database.DB, c Contributor) (Contributor, error) {
	validErr := c.Validate()
	if validErr != nil {
		return c, validErr
	}

	const query = "UPDATE " + t_contributors + " SET " +
		Cc_name + " = ?, " + Cc_biography + " = ? WHERE " + Cc_id + " = ? LIMIT 1"

	_, err := db.Exec(query, c.Name, c.Biography, c.Id)
	return c, err
}

func DeleteContributor(db database.DB, c Contributor) (Contributor, error) {
	const query = "DELETE FROM " + t_contributors + " WHERE " +
		Cc_id + " = ? LIMIT 1"
	_, err := db.Exec(query, c.Id)
	return c, err
}

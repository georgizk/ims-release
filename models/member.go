package models

import (
	"errors"
	"ims-release/database"
	"time"
)

type Member struct {
	Id        uint32    `json:"id"`
	Name      string    `json:"name"`
	Biography string    `json:"biography"`
	CreatedAt time.Time `json:"createdAt"`
}

const (
	t_members     string = "`members`"
	Mc_id         string = "`id`"
	Mc_name       string = "`name`"
	Mc_biography  string = "`biography`"
	Mc_created_at string = "`created_at`"

	Mmax_len_name      = 65535
	Mmax_len_biography = 65535
)

var (
	ErrNoSuchMember = errors.New("Could not find member.")
)

func NewMember(name, biography string, tm time.Time) Member {
	return Member{
		0,
		name,
		biography,
		tm,
	}
}

func FindMember(db database.DB, id uint32) (Member, error) {
	m := Member{}
	const query = "SELECT " + Mc_name + ", " + Mc_biography + ", " +
		Mc_created_at + " FROM " + t_members + " WHERE " + Mc_id + " = ?"

	row := db.QueryRow(query, id)
	err := row.Scan(&m.Name, &m.Biography, &m.CreatedAt)
	if err == database.ErrNoRows {
		return Member{}, ErrNoSuchMember
	} else if err != nil {
		return Member{}, err
	}
	m.Id = id
	return m, nil
}

func ListMembers(db database.DB) ([]Member, error) {
	members := []Member{}

	const query = "SELECT " + Mc_id + ", " + Mc_name + ", " +
		Mc_biography + ", " + Mc_created_at + " FROM " + t_members

	rows, err := db.Query(query)
	if err != nil {
		return []Member{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		err = rows.Scan(&m.Id, &m.Name, &m.Biography, &m.CreatedAt)
		if err != nil {
			return members, err
		}

		members = append(members, m)
	}
	err = rows.Err()
	return members, err
}

func (p *Member) Validate() error {
	if len(p.Name) > Mmax_len_name || len(p.Biography) > Mmax_len_biography {
		return ErrFieldTooLong
	}

	return nil
}

func SaveMember(db database.DB, m Member) (Member, error) {
	validErr := m.Validate()
	if validErr != nil {
		return m, validErr
	}

	const query = "INSERT INTO " + t_members + " (" +
		Mc_name + ", " + Mc_biography + ", " + Mc_created_at + ") VALUES (?, ?, ?, ?, ?)"

	res, err := db.Exec(query, m.Name, m.Biography, m.CreatedAt)
	if err != nil {
		return m, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return m, err
	}
	m.Id = uint32(id)
	return m, nil
}

func UpdateMember(db database.DB, m Member) (Member, error) {
	validErr := m.Validate()
	if validErr != nil {
		return m, validErr
	}

	const query = "UPDATE " + t_members + " SET " +
		Mc_name + " = ?, " + Mc_biography + " = ? WHERE " + Mc_id + " = ? LIMIT 1"

	_, err := db.Exec(query, m.Name, m.Biography, m.Id)
	return m, err
}

func DeleteMember(db database.DB, m Member) (Member, error) {
	const query = "DELETE FROM " + t_members + " WHERE " +
		Mc_id + " = ? LIMIT 1"
	_, err := db.Exec(query, m.Id)
	return m, err
}

package dao

import (
	"context"

	entity "rakit-tiket-be/pkg/entity/app_auth"

	"gitlab.com/threetopia/sqlgo/v2"
)

type UserDAO interface {
	Search(ctx context.Context, query entity.UserQuery) (entity.UsersEntity, error)
}

type userDAO struct {
	dbTrx DBTransaction
}

func MakeUserDAO(dbTrx DBTransaction) UserDAO {
	return userDAO{
		dbTrx: dbTrx,
	}
}

func (d userDAO) Search(ctx context.Context, query entity.UserQuery) (entity.UsersEntity, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("u.id", "id").
		SetSQLSelect("u.name", "name").
		SetSQLSelect("u.email", "email").
		SetSQLSelect("u.password_hash", "password_hash").
		SetSQLSelect("u.role", "role").
		SetSQLSelect("u.created_at", "created_at").
		SetSQLSelect("u.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom(`"user"`, "u")

	sqlWhere := sqlgo.NewSQLGoWhere()

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "u.id", "IN", query.IDs.Strings())
	}
	if len(query.Emails) > 0 {
		sqlWhere.SetSQLWhere("AND", "u.email", "IN", query.Emails)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoWhere(sqlWhere)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(
		ctx,
		sql.BuildSQL(),
		sql.GetSQLGoParameter().GetSQLParameter()...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users entity.UsersEntity
	for rows.Next() {
		var user entity.UserEntity
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.DaoEntity.CreatedAt,
			&user.DaoEntity.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

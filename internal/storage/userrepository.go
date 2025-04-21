package storage

import (
	"avito/internal/domain"
	postgresql "avito/pkg/database"
	"context"
	"errors"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type userRepositoryImpl struct {
	client postgresql.Client
}

func NewUserRepository(client postgresql.Client) domain.UserRepository {
	return userRepositoryImpl{client: client}
}

const selectUserBaseQuery = ` 
	   select
		  u.id as user_id
		, u.email as user_email
		, u.password as user_password
		, u.user_role_id as user_role_id
		, ur.name as user_role_name
   from users as u
   join user_roles as ur on ur.id = u.user_role_id
	`

// todo: what if there is no such user?
func scanUserFromRow(row pgx.Row) (user domain.User, err error) {
	err = row.Scan(&user.ID, &user.Email, &user.Password, &user.UserRole.ID, &user.UserRole.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			log.Println(pgErr)
			return user, errors.New("an error occured while user fetching from database")
		} else if err.Error() == "no rows in result set" {
			err = errors.New(domain.UserDoesNotExistsError)
			return
		}
		return
	}

	return user, nil
}

func (r userRepositoryImpl) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	const query string = selectUserBaseQuery + " where u.id = $1;"
	row := r.client.QueryRow(ctx, query, id)

	return scanUserFromRow(row)
}

func (r userRepositoryImpl) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	const query string = selectUserBaseQuery + " where email = $1;"
	row := r.client.QueryRow(ctx, query, email)

	return scanUserFromRow(row)
}

func (r userRepositoryImpl) Add(ctx context.Context, user domain.User) error {
	const query string = "insert into users (id, user_role_id, email, password) values ($1, $2, $3, $4);"

	_, err := r.client.Exec(ctx, query, user.ID, user.UserRole.ID, user.Email, user.Password)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			log.Println(pgErr)
			return errors.New("could not save user")
		}

		return err
	}

	return nil
}

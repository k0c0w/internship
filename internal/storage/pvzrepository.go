package storage

import (
	"avito/internal/domain"
	postgresql "avito/pkg/database"
	"context"
	"errors"
	"log"

	"github.com/jackc/pgconn"
)

type pvzRepositoryImpl struct {
	client postgresql.Client
}

func NewPVZRepository(client postgresql.Client) domain.PVZRepository {
	return &pvzRepositoryImpl{client: client}
}

func (r *pvzRepositoryImpl) FindById(ctx context.Context, id domain.PVZID) (domain.PVZ, error) {
	const query string = `
	select
			  p.id as id
			, p.creation_time_utc as creation_time_utc
			, c.id as city_id
			, c.name as city_name
	  from pvzs p
	  join cities c on c.id = p.city_id
	 where p.id = $1;
	`

	row := r.client.QueryRow(ctx, query, id)

	var pvz domain.PVZ
	err := row.Scan(&pvz.ID, &pvz.CreationTimeUTC, &pvz.City.ID, &pvz.City.Name)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return domain.PVZ{}, errors.New(domain.PVZDoesNotExistError)
		} else {
			return domain.PVZ{}, err
		}
	}

	return pvz, err
}

func (r pvzRepositoryImpl) Add(ctx context.Context, pvz domain.PVZ) error {
	const query string = "insert into pvzs (id, creation_time_utc, city_id) values ($1, $2, $3);"

	_, err := r.client.Exec(ctx, query, pvz.ID, pvz.CreationTimeUTC, pvz.City.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			log.Println(pgErr)
			return errors.New("could not save PVZ")
		}

		return err
	}

	return nil
}

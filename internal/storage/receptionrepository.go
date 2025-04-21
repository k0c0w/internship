package storage

import (
	"avito/internal/domain"
	postgresql "avito/pkg/database"
	"context"
	"fmt"
)

type receptionInfoRepositoryImpl struct {
	client postgresql.Client
}

func NewReceptionInfoRepository(client postgresql.Client) domain.ReceptionInfoRepository {
	return receptionInfoRepositoryImpl{
		client: client,
	}
}

func (r receptionInfoRepositoryImpl) Add(ctx context.Context, reception domain.ReceptionInfo) (err error) {
	const query string = "insert into receptions(id, pvz_id, creation_time_utc, status) values($1, $2, $3, $4);"

	_, err = r.client.Exec(ctx, query, reception.ID, reception.PVZID, reception.CreationTimeUTC, reception.Status)

	return
}

func (r receptionInfoRepositoryImpl) FindAllByFilter(ctx context.Context, filter domain.SearchReceptionInfoFilter) ([]domain.ReceptionInfo, error) {
	const queryBase string = `
	select 
			 id
		   , pvz_id
		   , creation_time_utc
		   , status
	  from receptions
	 where status = $1 and pvz_id = $2
     order by creation_time_utc %s
	 limit $3;
	`
	ordering := "asc"
	if filter.DescendingDateOrdering {
		ordering = "desc"
	}

	query := fmt.Sprintf(queryBase, ordering)

	rows, err := r.client.Query(ctx, query, filter.Status, filter.PVZID, filter.Limit)
	if err != nil {
		return nil, err
	}

	receptions := make([]domain.ReceptionInfo, 0)

	for rows.Next() {
		var reception domain.ReceptionInfo

		err := rows.Scan(&reception.ID, &reception.PVZID, &reception.CreationTimeUTC, &reception.Status)
		if err != nil {
			return nil, err
		} else {
			receptions = append(receptions, reception)
		}
	}

	return receptions, nil
}

func (r receptionInfoRepositoryImpl) Update(ctx context.Context, reception domain.ReceptionInfo) error {
	const query string = `
	update receptions
	   set 
	   	    pvz_id = $2
	   	  , creation_time_utc = $3
		  , status = $4
	 where id = $1
	`
	_, err := r.client.Exec(ctx, query, reception.ID, reception.PVZID, reception.CreationTimeUTC, reception.Status)
	return err
}

package storage

import (
	"avito/internal/domain"
	postgresql "avito/pkg/database"
	"context"
	"encoding/json"
	"fmt"
)

type pvzReportRepositoryImpl struct {
	client postgresql.Client
}

func (p pvzReportRepositoryImpl) FindAllByFilter(ctx context.Context, filter domain.SearchPVZReportAggregateFilter) ([]*domain.PVZReportAggregate, error) {
	const queryFormat string = `
		select 
			  p.id
		    , p.creation_time_utc
		    , c.id as city_id
			, c.name as city_name
		    , coalesce((
            	select array_to_json(array_agg(row_to_json(r)))
            	  from (
                	 select 
                    		(select row_to_json(i)
                       		   from (select 
                        	        r.id
                        	        , r.creation_time_utc
                        	        , r.status
                        	        , r.pvz_id
                        	   ) i
                    		) as info
                    		, r.products
                	    from receptions_with_products_view r
                   		join pvzs pr on pr.id = r.pvz_id
                  	   where r.pvz_id = p.id %s
            ) r
		        ),
		        '[]'::json
		     ) as receptions
		  from pvzs p
		  join cities c on p.city_id = c.id
		 where p.pvz_record_number > $1
		 limit $2;
		`

	arguments := make([]interface{}, 0, 4)
	offset := (filter.Page - 1) * filter.Limit
	arguments = append(arguments, offset, filter.Limit)

	var receptionFilterQuery string = ""
	if filter.ReceptionEndTimeUTC != nil && filter.ReceptionStartTimeUTC != nil {
		receptionFilterQuery = " and (r.creation_time_utc between $3 and $4) "
		arguments = append(arguments, *filter.ReceptionStartTimeUTC, *filter.ReceptionEndTimeUTC)
	} else if filter.ReceptionStartTimeUTC != nil {
		receptionFilterQuery = " and (r.creation_time_utc >= $3)"
		arguments = append(arguments, *filter.ReceptionStartTimeUTC)
	} else if filter.ReceptionEndTimeUTC != nil {
		receptionFilterQuery = " and (r.creation_time_utc < $3)"
		arguments = append(arguments, *filter.ReceptionEndTimeUTC)
	}

	query := fmt.Sprintf(queryFormat, receptionFilterQuery)
	rows, err := p.client.Query(ctx, query, arguments...)

	if err != nil {
		return nil, err
	}

	reports := make([]*domain.PVZReportAggregate, 0)
	for rows.Next() {
		var receptions []domain.ReceptionAggregate
		var pvz domain.PVZ
		var receptionsJSON []byte

		err := rows.Scan(&pvz.ID, &pvz.CreationTimeUTC, &pvz.City.ID, &pvz.City.Name, &receptionsJSON)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(receptionsJSON, &receptions); err != nil {
			return nil, err
		}

		report := domain.PVZReportAggregate{
			PVZ:        &pvz,
			Receptions: receptions,
		}

		reports = append(reports, &report)
	}

	return reports, nil
}

func NewPVZReportAggregateRepository(client postgresql.Client) domain.PVZReportAggregateRepository {
	return pvzReportRepositoryImpl{client: client}
}

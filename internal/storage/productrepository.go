package storage

import (
	"avito/internal/domain"
	postgresql "avito/pkg/database"
	"context"
)

type productRepositoryImpl struct {
	client postgresql.Client
}

func NewProductRepository(client postgresql.Client) domain.ProductRepository {
	return productRepositoryImpl{client: client}
}

func (p productRepositoryImpl) Add(ctx context.Context, product domain.Product) error {
	const query string = "insert into products(id, reception_id, creation_time_utc, category) values($1, $2, $3, $4);"

	_, err := p.client.Exec(ctx, query, product.ID, product.ReceptionID, product.CreationTimeUTC, product.Category)

	return err
}

func (p productRepositoryImpl) FindAllByReceptionID(ctx context.Context, receptionId domain.ReceptionID) ([]*domain.Product, error) {
	const query string = `
		select
				  id
				, reception_id  
				, creation_time_utc
				, category
		  from products
		 where reception_id = $1;
	`

	rows, err := p.client.Query(ctx, query, receptionId)

	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, 0)
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(&product.ID, &product.ReceptionID, &product.CreationTimeUTC, &product.Category)
		if err != nil {
			return nil, err
		}

		products = append(products, &product)
	}

	return products, nil
}

func (p productRepositoryImpl) Remove(ctx context.Context, product domain.Product) error {
	const query string = "delete from products where id = $1;"

	_, err := p.client.Exec(ctx, query, product.ID)

	return err
}

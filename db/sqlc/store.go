package db

import (
	"context"
	"database/sql"
)

type Store struct {
	*Queries         // each query only do 1 operation on 1 specific table. So queries struct doesn't transaction so we embeddig queries all individuals query functions provided by queries will be available to store
	db       *sql.DB // becuase store required to create a new db transaction
}

func NewStore(db *sql.DB) *Store { //use sql.DB as input
	return &Store{
		db:      db,
		Queries: New(db), // queries is created by callint the New() , New function is generated by sqlc it creates and returns a Queries

	}
}

func (store *Store) execTx(ctx context.Context,fn func(*Queries) error) error{ // takes a context and a callback function as input, then it will start a new db transaction
	tx,err := store.db.BeginTx((ctx, &sql.TxOptions{})) //start a new transaction we call store.db.BeginTx() pass in the context, and optionally a sql.TxOptions
	// this option allow us to set a custom isolation level for this transaction
	if err != nil {
		return err
	}
	q := New(tx) // we passing to tx intead of passing to db
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v",err,rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct{
	FromAccountID int64 'json:"from_account_id'
	ToAccountID int64 'json:"to_account_id"'
	Amount int64 'json:"amount"'
}

type TransferTxResult struct {
    Transfer    Transfer `json:"transfer"`
    FromAccount Account  `json:"from_account"`
    ToAccount   Account  `json:"to_account"`
    FromEntry   Entry    `json:"from_entry"` // records that money is moving out of the FromAcc 
    ToEntry     Entry    `json:"to_entry"` //records that money is moving in of the FromAcc 
}

func (store *Store) TransfereTx(ctx context.Context, arg TransferTxParams){
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
            FromAccountID: arg.FromAccountID,
            ToAccountID:   arg.ToAccountID,
            Amount:        arg.Amount,

	})
	if err != nil{
	return err 

	}
	return nil
})
return result,err
}

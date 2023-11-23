package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

type Repository struct {
	client *goredislib.Client
	mutex  *redsync.Mutex
	//mutex  *sync.Mutex
}

func NewRepository(address string) Repository {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: address,
	})

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	mutexname := "my-global-mutex"

	/*
		mutex := sync.Mutex{}
	*/
	return Repository{
		client: client,
		mutex:  rs.NewMutex(mutexname),
		//mutex: sync.Mutex{},
	}
}

func (r *Repository) BuySharesWithRedisLock(ctx context.Context, userId, companyId string, numShares int) error {

	/*
		r.mutex.Lock()
		defer r.mutex.Unlock()
	*/
	if err := r.mutex.Lock(); err != nil {
		fmt.Printf("error during lock: %v \n", err)
	}

	defer func() {
		if ok, err := r.mutex.Unlock(); !ok || err != nil {
			fmt.Printf("error during unlock: %v \n", err)
		}
	}()

	currentShares, err := r.client.Get(ctx, BuildCompanySharesKey(companyId)).Int()
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	if currentShares < numShares {
		fmt.Print("error: company does not have enough shares \n")
		return errors.New("error: company does not have enough shares")
	}
	currentShares -= numShares

	r.client.Set(ctx, BuildCompanySharesKey(companyId), currentShares, 0)

	return nil
}

func (r *Repository) GetCompanyShares(ctx context.Context, companyId string) (int, error) {
	result := r.client.Get(ctx, BuildCompanySharesKey(companyId))
	currentShares, err := result.Int()
	if err != nil {
		return 0, err
	}
	return currentShares, nil
}

func (r *Repository) PublishShares(ctx context.Context, companyId string, numShares int) error {
	status := r.client.Set(ctx, BuildCompanySharesKey(companyId), numShares, 0)
	return status.Err()
}

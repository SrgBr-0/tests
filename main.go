package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type Repository interface {
	UpdateBalanceWithPayment(ctx context.Context, uid int, paymentAmount int) error
}

type repository struct {
	db Database
}

func NewRepository(db Database) Repository {
	return &repository{db}
}

type Database interface {
	Conn() (*sql.DB, error)
	ConnWith(ctx context.Context) (*sql.DB, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func (r *repository) UpdateBalanceWithPayment(ctx context.Context, uid int, paymentAmount int) error {
	tx, err := r.db.ConnWith(ctx)
	if err != nil {
		return err
	}
	defer tx.Close()

	_, err = tx.ExecContext(ctx, "BEGIN")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Обновляем баланс
	_, err = tx.ExecContext(ctx, "UPDATE userbalance SET balance = balance - $1 WHERE id = $2 AND (balance - $1) >= 0", paymentAmount, uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Успешно завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

type FakeDB struct {
	userBalances map[int]int
}

func NewFakeDB() *FakeDB {
	return &FakeDB{
		userBalances: make(map[int]int),
	}
}

func (f *FakeDB) Conn() (*sql.DB, error) {
	db, _, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (f *FakeDB) ConnWith(ctx context.Context) (*sql.DB, error) {
	db, _, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (f *FakeDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (f *FakeDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (f *FakeDB) GetBalance(userID int) int {
	return f.userBalances[userID]
}

func (f *FakeDB) UpdateBalance(userID, amount int) {
	f.userBalances[userID] += amount
}

func TestUpdateBalanceWithPayment(t *testing.T) {
	// Создаем экземпляр FakeDB
	fakeDB := NewFakeDB()
	repositoryInstance := NewRepository(fakeDB)

	// Создаем тестовый контекст
	ctx := context.Background()

	// Добавляем тестового пользователя в фейковую базу данных
	fakeDB.userBalances[1] = 100

	// Тестовые данные
	testUserID := 1
	testPaymentAmount := 50

	// Вызываем функцию UpdateBalanceWithPayment с тестовыми данными
	err := repositoryInstance.UpdateBalanceWithPayment(ctx, testUserID, testPaymentAmount)

	// Проверяем результат
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Проверяем изменения в фейковой базе данных (fakeDB)
	if updatedBalance := fakeDB.GetBalance(testUserID); updatedBalance != 50 {
		t.Fatalf("Expected updated balance to be 50, got %v", updatedBalance)
	}
}

func main() {
	// Создаем экземпляр FakeDB
	fakeDB := NewFakeDB()
	repositoryInstance := NewRepository(fakeDB)

	// Создаем тестовый контекст
	ctx := context.Background()

	// Добавляем тестового пользователя в фейковую базу данных
	fakeDB.userBalances[1] = 100

	// Тестовые данные
	testUserID := 1
	testPaymentAmount := 50

	// Вызываем функцию UpdateBalanceWithPayment с тестовыми данными
	err := repositoryInstance.UpdateBalanceWithPayment(ctx, testUserID, testPaymentAmount)

	// Проверяем результат
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Balance updated successfully")
	}
}

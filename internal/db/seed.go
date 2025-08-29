package db

// import (
// 	"context"
// 	"social/internal/store"
// 	"strconv"
// )

// func Seed(store store.Storage) error {
// 	ctx := context.Background()

// 	users := generateUsers(100)

// 	return nil
// }

// func generateUsers(n int) []*store.User {
// 	users := make([]*store.User, n)
// 	for i := 0; i < n; i++ {
// 		i2 := i + i + i + i
// 		s := strconv.Itoa(i2)
// 		users[i] = &store.User{
// 			Username: s,
// 			Email:    (s + "@email.com"),
// 			Password: s + s + s + s,
// 		}
// 	}

// 	return users
// }

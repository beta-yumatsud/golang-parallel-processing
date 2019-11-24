// +build integration
// こう書くと「go test -tags=integrationで実行される

package main

func sum(a, b int) int {
	return a + b
}

// reflect.DeepEqualとか便利

// go test -cover でカバレッジ表示
// go test -coverprofile=hogehoge.out
// go tool cover -html=hogehoge.out
// go test -covermode=countで各ステートメントの実行回数が記載される
